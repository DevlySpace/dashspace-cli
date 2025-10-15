var require = function(name) {
    if (name === 'react') return React;
    if (name === 'react-dom') return ReactDOM;
    if (name === 'dashspace-lib') {
        return {
            Provider: {
                GITHUB: 'github',
                STRIPE: 'stripe',
                GITLAB: 'gitlab',
                LINEAR: 'linear',
                SHOPIFY: 'shopify',
                GOOGLE: 'google',
                SENTRY: 'sentry',
                VERCEL: 'vercel',
                NETLIFY: 'netlify',
                ATLASSIAN: 'atlassian',
                ASANA: 'asana',
                PAGERDUTY: 'pagerduty',
                SLACK: 'slack',
                DISCORD: 'discord',
                NOTION: 'notion',
                CALENDLY: 'calendly',
                AIRTABLE: 'airtable',
                FIGMA: 'figma'
            },

            ModuleInterfaces: {
                ISearchable: 'ISearchable',
                IRefreshable: 'IRefreshable',
                IExportable: 'IExportable',
                IFilterable: 'IFilterable',
                IThemeable: 'IThemeable',
                INotifiable: 'INotifiable',
                IDataProvider: 'IDataProvider',
                ISchedulable: 'ISchedulable'
            },

            useModuleConfig: context.hooks.useModuleConfig,
            useProvider: context.hooks.useProvider,
            useModuleStorage: context.hooks.useModuleStorage,
            useModuleEvent: context.hooks.useModuleEvent,
            useNotification: context.hooks.useNotification,
            useModuleInterfaces: context.hooks.useModuleInterfaces,
            useWebhookEvents: context.hooks.useWebhookEvents,
            useProviderWebhooks: context.hooks.useProviderWebhooks,
            useDataProvider: context.hooks.useDataProvider,
            useDataQuery: context.hooks.useDataQuery,

            dataRegistry: (typeof window !== 'undefined' && window.parent && window.parent.DashspaceLib && window.parent.DashspaceLib.dataRegistry)
                ? window.parent.DashspaceLib.dataRegistry
                : (typeof window !== 'undefined' && window.DashspaceLib && window.DashspaceLib.dataRegistry)
                    ? window.DashspaceLib.dataRegistry
                    : undefined,

            UQLParser: (typeof window !== 'undefined' && window.parent && window.parent.DashspaceLib && window.parent.DashspaceLib.UQLParser)
                ? window.parent.DashspaceLib.UQLParser
                : (typeof window !== 'undefined' && window.DashspaceLib && window.DashspaceLib.UQLParser)
                    ? window.DashspaceLib.UQLParser
                    : undefined,

            UQLValidator: (typeof window !== 'undefined' && window.parent && window.parent.DashspaceLib && window.parent.DashspaceLib.UQLValidator)
                ? window.parent.DashspaceLib.UQLValidator
                : (typeof window !== 'undefined' && window.DashspaceLib && window.DashspaceLib.UQLValidator)
                    ? window.DashspaceLib.UQLValidator
                    : undefined,

            UQL_EXAMPLES: (typeof window !== 'undefined' && window.parent && window.parent.DashspaceLib && window.parent.DashspaceLib.UQL_EXAMPLES)
                ? window.parent.DashspaceLib.UQL_EXAMPLES
                : (typeof window !== 'undefined' && window.DashspaceLib && window.DashspaceLib.UQL_EXAMPLES)
                    ? window.DashspaceLib.UQL_EXAMPLES
                    : [],

            BaseModule: function(contextArg, metadata, configurationSteps, providers) {
                this.context = contextArg;
                this.metadata = metadata;
                this.configurationSteps = configurationSteps || [];
                this.providers = providers || [];
                this.config = contextArg.config.get() || {};
                this.initialized = false;
                this.cleanupHandlers = [];
                this.webhookHandlers = new Map();

                var self = this;

                this.getConfig = function() {
                    return self.config;
                };

                this.getMetadata = function() {
                    return self.metadata;
                };

                this.getConfigurationSteps = function() {
                    return self.configurationSteps;
                };

                this.getRequiredProviders = function() {
                    return self.providers;
                };

                this.callProvider = function(providerName, options) {
                    return contextArg.providers.call(providerName, options);
                };

                this.storeData = function(key, value) {
                    return contextArg.storage.set(key, value);
                };

                this.getStoredData = function(key) {
                    return contextArg.storage.get(key);
                };

                this.removeStoredData = function(key) {
                    return contextArg.storage.remove(key);
                };

                this.emit = function(event, data) {
                    if (contextArg.events && contextArg.events.emit) {
                        contextArg.events.emit(event, data);
                    }
                };

                this.on = function(event, callback) {
                    if (contextArg.events && contextArg.events.on) {
                        var unsubscribe = contextArg.events.on(event, callback);
                        self.cleanupHandlers.push(unsubscribe);
                        return unsubscribe;
                    }
                    return function() {};
                };

                this.showNotification = function(message, type, duration) {
                    if (contextArg.ui && contextArg.ui.showNotification) {
                        contextArg.ui.showNotification(message, type, duration);
                    }
                };

                this.registerWebhookHandler = function(eventType, handler) {
                    self.webhookHandlers.set(eventType, handler);
                    console.log('[BaseModule] Registered webhook handler for:', eventType);
                };

                this.unregisterWebhookHandler = function(eventType) {
                    self.webhookHandlers.delete(eventType);
                };

                this.handleWebhookEvent = function(event) {
                    var handler = self.webhookHandlers.get(event.type) ||
                        self.webhookHandlers.get(event.provider);

                    if (handler) {
                        try {
                            Promise.resolve(handler(event)).catch(function(error) {
                                console.error('[BaseModule] Webhook handler error:', error);
                                self.emit('webhook:error', { event: event, error: error });
                            });
                        } catch (error) {
                            console.error('[BaseModule] Webhook handler error:', error);
                            self.emit('webhook:error', { event: event, error: error });
                        }
                    } else {
                        console.warn('[BaseModule] No handler for webhook event:', event.type);
                    }
                };

                this.getWebhookConfig = function() {
                    if (!self.metadata.webhooks) return null;

                    var configValues = {};
                    var configFields = self.metadata.webhooks.configFields || [];

                    for (var i = 0; i < configFields.length; i++) {
                        var field = configFields[i];
                        if (self.config[field]) {
                            configValues[field] = self.config[field];
                        }
                    }

                    return {
                        provider: self.metadata.webhooks.provider,
                        events: self.metadata.webhooks.events,
                        metadata: configValues
                    };
                };

                this.initialize = function() {
                    return Promise.resolve();
                };

                this.cleanup = function() {
                    return Promise.resolve();
                };

                this.onConfigChange = function(newConfig, oldConfig) {
                };

                this.start = function() {
                    if (self.initialized) return Promise.resolve();
                    return self.initialize().then(function() {
                        self.initialized = true;
                    });
                };

                this.stop = function() {
                    if (!self.initialized) return Promise.resolve();
                    return self.cleanup().then(function() {
                        self.cleanupHandlers.forEach(function(handler) {
                            try {
                                handler();
                            } catch (e) {
                                console.error('Cleanup error:', e);
                            }
                        });
                        self.cleanupHandlers = [];
                        self.webhookHandlers.clear();
                        self.initialized = false;
                    });
                };

                if (metadata.webhooks && metadata.webhooks.handler) {
                    this.registerWebhookHandler(metadata.webhooks.provider, metadata.webhooks.handler);
                }
            },

            ConfigurationStep: function(options) {
                this.id = options.id;
                this.title = options.title;
                this.description = options.description;
                this.fields = options.fields || [];
                this.order = options.order || 0;
                this.optional = options.optional || false;
            },

            TextField: function(options) {
                this.type = 'text';
                this.name = options.name;
                this.label = options.label;
                this.description = options.description;
                this.defaultValue = options.defaultValue;
                this.validation = options.validation || {};
                this.placeholder = options.placeholder;
            },

            NumberField: function(options) {
                this.type = 'number';
                this.name = options.name;
                this.label = options.label;
                this.description = options.description;
                this.defaultValue = options.defaultValue;
                this.validation = options.validation || {};
                this.min = options.min;
                this.max = options.max;
                this.step = options.step;
            },

            SelectField: function(options) {
                this.type = 'select';
                this.name = options.name;
                this.label = options.label;
                this.description = options.description;
                this.defaultValue = options.defaultValue;
                this.validation = options.validation || {};
                this.options = options.options || [];
                this.multiple = options.multiple || false;
            },

            BooleanField: function(options) {
                this.type = 'boolean';
                this.name = options.name;
                this.label = options.label;
                this.description = options.description;
                this.defaultValue = options.defaultValue || false;
                this.validation = options.validation || {};
            },

            PasswordField: function(options) {
                this.type = 'password';
                this.name = options.name;
                this.label = options.label;
                this.description = options.description;
                this.defaultValue = options.defaultValue;
                this.validation = options.validation || {};
                this.placeholder = options.placeholder;
            },

            UrlField: function(options) {
                this.type = 'url';
                this.name = options.name;
                this.label = options.label;
                this.description = options.description;
                this.defaultValue = options.defaultValue;
                this.validation = options.validation || {};
                this.placeholder = options.placeholder;
            }
        };
    }

    console.warn('Unknown dependency requested:', name);
    return null;
};