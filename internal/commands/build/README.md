# Build System Architecture

## Overview

The build system compiles TypeScript/React modules into JavaScript bundles executable in the Dashspace environment. It handles metadata extraction, validation, compilation, and manifest generation.

## Build Process Steps

The build command executes the following steps in order:

### 1. Structure Validation
Validates the basic project structure:
- Checks for `Module.ts` or `Module.tsx` (required)
- Checks for `Component.tsx` (warns if missing)
- Validates `package.json` existence
- Ensures proper project layout

### 2. Dependency Installation
- Checks if `node_modules` exists
- Runs `npm install` if dependencies are missing
- Ensures all required packages are available

### 3. TypeScript Validation (--strict mode, enabled by default)
Comprehensive TypeScript checking:
- **Type Checking**: Runs `tsc --noEmit` to validate all TypeScript types
- **dashspace-lib Compatibility**: Ensures dashspace-lib types are properly used
- **Interface Implementation**: Validates that declared interfaces are correctly implemented
- **Unused Code Detection**: Identifies unused imports and variables
- **Type Safety Patterns**: Checks for proper use of `satisfies` operator

### 4. Linting (--strict mode, enabled by default)
Code quality and style checking:
- **ESLint Validation**: Runs ESLint with TypeScript rules
- **React Best Practices**: Validates React hooks usage and patterns
- **Package.json Validation**: Checks dependency categorization (dependencies vs devDependencies)
- **Common Issues Detection**:
    - Console.log statements in production code
    - Hardcoded URLs (localhost, 127.0.0.1)
    - TODO/FIXME comments
    - Usage of `any` type
    - Missing error handling
    - Missing loading states

### 5. Metadata Extraction
Parses Module.ts to extract:
- Module ID (required)
- Module name (required)
- Module version (required)
- Module slug (required)
- Description, author, icon, category, tags
- Configuration steps (if any)
- Required providers with scopes
- Implemented interfaces
- Required permissions

### 6. Compilation
TypeScript to JavaScript compilation using esbuild:
- Bundles all code into a single file
- Transforms JSX/TSX to JavaScript
- Tree shaking for optimal bundle size
- Minification in production mode
- External dependencies (react, react-dom, dashspace-lib) are not bundled
- Target: ES2020
- Format: IIFE (Immediately Invoked Function Expression)

### 7. Bundle Generation
Wraps the compiled code:
- Adds module loader wrapper
- Injects dashspace-lib runtime polyfill
- Creates module initialization function
- Generates SHA256 checksum for integrity

### 8. Output Validation (--strict mode only)
Final checks on generated files:
- Verifies `bundle.js` exists and is valid
- Validates `dashspace.json` structure and required fields
- Checks bundle size (warns if >500KB)
- Ensures manifest completeness

## File Structure

```
build/
├── build.go         # Entry point and orchestration
├── parser.go        # Data extraction from Module.ts
├── interfaces.go    # Interface implementation validation
├── compiler.go      # TypeScript compilation via esbuild
├── generator.go     # Bundle generation and wrapping
├── validator.go     # Project structure validation
├── typescript.go    # TypeScript-specific validation
├── linting.go       # ESLint and code quality checks
├── writer.go        # Output file writing
├── watcher.go       # Watch mode for development
├── extractor.go     # Specialized extractors (steps, providers)
├── types.go         # Shared type definitions
└── utils.go         # Common utility functions
```

## Command Usage

### Basic Commands

```bash
# Standard build with all validations (recommended)
dashspace build

# Development build without minification
dashspace build --dev

# Watch mode for development
dashspace build --watch

# Build without strict mode (warnings don't fail)
dashspace build --no-strict

# Quick build skipping validation (not recommended)
dashspace build --skip-checks

# Custom output directory
dashspace build -o ./my-dist
```

### Build Modes

#### Strict Mode (Default)
- All warnings are treated as errors
- Full TypeScript validation
- Full ESLint validation
- Output validation
- Recommended for production builds

#### Non-Strict Mode (--no-strict)
- Warnings are displayed but don't fail the build
- Useful during development
- Not recommended for production

#### Skip Checks Mode (--skip-checks)
- Bypasses TypeScript and linting validation
- Faster builds but risky
- Should only be used when you're certain the code is valid

## Output Structure

### bundle.js
- Minified (production) or readable (development)
- Wrapped in module loader
- Includes dashspace-lib runtime polyfill
- External dependencies not bundled

### dashspace.json
```json
{
  "id": 1,
  "name": "module-name",
  "version": "1.0.0",
  "slug": "module-slug",
  "description": "Module description",
  "author": "Author name",
  "entry": "bundle.js",
  "checksum": "sha256-hash",
  "timestamp": "ISO-8601",
  "requires_setup": false,
  "implemented_interfaces": ["ISearchable", "IRefreshable"],
  "providers": [...],
  "configuration_steps": [...],
  "permissions": [...],
  "build_info": {
    "cli_version": "1.0.10",
    "build_date": "ISO-8601",
    "validated": true
  }
}
```

## TypeScript Configuration

The build system automatically creates a `tsconfig.json` if not present:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "jsx": "react",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "moduleResolution": "node",
    "resolveJsonModule": true,
    "noEmit": true,
    "types": ["react", "react-dom", "node"]
  },
  "include": ["**/*.ts", "**/*.tsx"],
  "exclude": ["node_modules", "dist", "build"]
}
```

## ESLint Configuration

The build system automatically creates `.eslintrc.json` if not present:

```json
{
  "parser": "@typescript-eslint/parser",
  "extends": [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended"
  ],
  "plugins": ["@typescript-eslint", "react", "react-hooks"],
  "parserOptions": {
    "ecmaVersion": 2020,
    "sourceType": "module",
    "ecmaFeatures": {
      "jsx": true
    }
  },
  "settings": {
    "react": {
      "version": "detect"
    }
  },
  "env": {
    "browser": true,
    "es2020": true,
    "node": true
  },
  "rules": {
    "react/react-in-jsx-scope": "off",
    "@typescript-eslint/no-explicit-any": "warn",
    "@typescript-eslint/explicit-module-boundary-types": "off",
    "no-console": ["warn", { "allow": ["warn", "error"] }],
    "react/prop-types": "off"
  },
  "ignorePatterns": ["dist/", "build/", "node_modules/", "*.js"]
}
```

## Interface Validation

The build system validates that interfaces declared in Module.ts are properly implemented in Component.tsx:

### Required Interface Methods

```typescript
interface ISearchable {
  search(): void
  getSearchResults(): any[]
  getSearchFilters(): any[]
}

interface IRefreshable {
  refresh(): void
  getLastRefresh(): Date
  setAutoRefresh(enabled: boolean): void
}

interface IExportable {
  export(): void
  getSupportedFormats(): string[]
  exportData(format: string): any
}

interface IFilterable {
  applyFilter(filter: string): void
  clearFilters(): void
    getCurrentFilter(): string
}
```

## Common Issues and Solutions

### "Module.ts not found"
Ensure Module.ts or Module.tsx exists in root or src/

### "No valid factory function"
Check that DashspaceModuleFactory is properly exported

### Interface validation fails
Ensure Component.tsx implements all declared interfaces with required methods

### Build size too large
- Check for accidentally bundled dependencies
- Ensure external dependencies are properly marked
- Enable tree shaking and minification

### TypeScript errors
- Run `npx tsc --noEmit` to see detailed errors
- Ensure all types are properly defined
- Check that dashspace-lib types are imported correctly

### ESLint warnings
- Run `npx eslint . --ext .ts,.tsx` to see all warnings
- Fix or suppress warnings as appropriate
- Use `--no-strict` during development if needed

## Performance Considerations

- Regex parsing is used for simplicity but could be replaced with AST parsing for accuracy
- File watching uses debouncing (300ms) to prevent excessive rebuilds
- Tree shaking and minification reduce bundle size
- TypeScript checking can be skipped with `--skip-checks` for faster builds during development

## Testing Checklist

- [ ] Module without Component.tsx builds with warning
- [ ] Module with interfaces validates implementation
- [ ] Configuration steps are correctly extracted
- [ ] Providers are properly identified
- [ ] Watch mode rebuilds on file changes
- [ ] Minification works in production mode
- [ ] TypeScript validation catches type errors
- [ ] ESLint validation catches code quality issues
- [ ] Bundle size warnings appear for large bundles

## Development Workflow

1. **Initial Setup**
   ```bash
   npm init
   npm install dashspace-lib react react-dom
   npm install -D typescript @types/react @types/react-dom eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin
   ```

2. **Development Mode**
   ```bash
   dashspace build --watch --dev --no-strict
   ```

3. **Pre-commit Check**
   ```bash
   dashspace build
   ```

4. **Production Build**
   ```bash
   dashspace build --strict
   ```

## Future Improvements

- [ ] AST parsing instead of regex for better accuracy
- [ ] Incremental compilation for faster rebuilds
- [ ] Better source map support
- [ ] Module dependency resolution
- [ ] Hot reload support for development
- [ ] Build profiles (development, staging, production)
- [ ] Configuration file support (.dashspace.yml)
- [ ] Parallel validation for faster builds