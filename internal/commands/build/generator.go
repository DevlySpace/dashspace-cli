package build

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
)

type Generator struct {
	config *DashspaceConfig
}

func NewGenerator(config *DashspaceConfig) *Generator {
	return &Generator{config: config}
}

func (g *Generator) Generate(moduleCode string) (string, string) {
	polyfillContent, err := g.loadPolyfillTemplate()
	if err != nil {
		fmt.Printf("⚠️  Warning: %v\n", err)
		polyfillContent = ""
	}

	wrappedCode := g.wrapModule(moduleCode, polyfillContent, g.config.ID)

	hash := sha256.Sum256([]byte(wrappedCode))
	checksum := hex.EncodeToString(hash[:])

	return wrappedCode, checksum
}

func (g *Generator) loadPolyfillTemplate() (string, error) {
	runtimePath := "/usr/local/share/dashspace/templates/dashspace-lib-runtime.js"
	if !fileExists(runtimePath) {
		return "", fmt.Errorf("dashspace-lib runtime not found at %s", runtimePath)
	}

	content, err := ioutil.ReadFile(runtimePath)
	if err != nil {
		return "", fmt.Errorf("failed to read dashspace-lib runtime: %w", err)
	}

	fmt.Printf("✅ Using dashspace-lib runtime: %s\n", runtimePath)
	return string(content), nil
}

func (g *Generator) wrapModule(moduleCode string, polyfillContent string, moduleID int) string {
	wrapper := `(function(global) {
    global.__module_%d = {
        init: function(context, deps) {
            var module = { exports: {} };
            var exports = module.exports;
            
            var React = deps.React;
            var ReactDOM = deps.ReactDOM;
            var useState = React.useState;
            var useEffect = React.useEffect;
            var useCallback = React.useCallback;
            var useMemo = React.useMemo;
            var useRef = React.useRef;
            var useContext = React.useContext;
            var useReducer = React.useReducer;
            var createElement = React.createElement;
            var Fragment = React.Fragment;
            
            %s
            
            var __webpack_require__ = require;
            var __webpack_exports__ = exports;
            
            (function(React, ReactDOM, require, module, exports) {
                %s
            })(React, ReactDOM, require, module, exports);
            
            var exportedFactory = null;
            
            if (module.exports && typeof module.exports === 'object') {
                if (typeof module.exports.default === 'function') {
                    exportedFactory = module.exports.default;
                } else if (module.exports.__esModule && typeof module.exports.default === 'function') {
                    exportedFactory = module.exports.default;
                } else {
                    for (var key in module.exports) {
                        if (typeof module.exports[key] === 'function' && key !== '__esModule') {
                            exportedFactory = module.exports[key];
                            break;
                        }
                    }
                }
            } else if (typeof module.exports === 'function') {
                exportedFactory = module.exports;
            }
            
            if (!exportedFactory || typeof exportedFactory !== 'function') {
                throw new Error('No valid factory function found in module exports');
            }
            
            var result = exportedFactory(context);
            
            if (!result || typeof result !== 'object') {
                throw new Error('Factory must return an object, got: ' + typeof result);
            }
            
            if (!result.Component) {
                throw new Error('Factory result missing Component property');
            }
            
            if (typeof result.Component !== 'function') {
                throw new Error('Component must be a function, got: ' + typeof result.Component);
            }
            
            return result;
        }
    };
})(window);`

	return fmt.Sprintf(wrapper, moduleID, polyfillContent, moduleCode)
}
