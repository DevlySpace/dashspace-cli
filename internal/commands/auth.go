package commands

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/devlyspace/dashspace-cli/internal/api"
	"github.com/devlyspace/dashspace-cli/internal/config"
	"github.com/spf13/cobra"
)

func handleWebLogin() error {
	fmt.Println("🚀 Opening DashSpace app for authentication...")

	state := generateRandomState()

	// Deep link to the app
	deepLink := fmt.Sprintf("dashspace://auth/cli?state=%s", state)

	fmt.Printf("🔗 If the app doesn't open automatically, click: %s\n", deepLink)
	fmt.Println("⏳ Waiting for authentication...")

	if err := openDeepLink(deepLink); err != nil {
		fmt.Printf("⚠️  Failed to open app: %v\n", err)
		fmt.Println("💡 Please open DashSpace app manually and go to Settings > CLI Authentication")
	}

	authResponse, err := waitForAuthCallback(state)
	if err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	return saveAuthConfig(authResponse)
}

func waitForAuthCallback(expectedState string) (*api.AuthResponse, error) {
	resultChan := make(chan *api.AuthResponse, 1)
	errorChan := make(chan error, 1)

	// Start temporary server to receive callback
	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		token := r.URL.Query().Get("token")
		username := r.URL.Query().Get("username")
		email := r.URL.Query().Get("email")

		if state != expectedState {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			errorChan <- fmt.Errorf("invalid state parameter")
			return
		}

		if token == "" {
			errorMessage := r.URL.Query().Get("error")
			if errorMessage == "" {
				errorMessage = "No token received"
			}
			http.Error(w, errorMessage, http.StatusBadRequest)
			errorChan <- fmt.Errorf("authentication failed: %s", errorMessage)
			return
		}

		// Success response
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
            <!DOCTYPE html>
            <html>
            <head>
                <title>DashSpace CLI - Success</title>
                <style>
                    body { 
                        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
                        text-align: center; padding: 50px; background: #f5f5f5;
                        display: flex; align-items: center; justify-content: center;
                        min-height: 100vh; margin: 0;
                    }
                    .container { 
                        background: white; padding: 40px; border-radius: 12px;
                        box-shadow: 0 4px 20px rgba(0,0,0,0.1); max-width: 400px;
                    }
                    .success { color: #28a745; font-size: 48px; margin-bottom: 20px; }
                    h1 { color: #333; margin-bottom: 10px; }
                    p { color: #666; margin-bottom: 15px; }
                    .close-notice { background: #e3f2fd; padding: 15px; border-radius: 8px; 
                                   color: #1976d2; font-size: 14px; margin-top: 20px; }
                </style>
            </head>
            <body>
                <div class="container">
                    <div class="success">✅</div>
                    <h1>Authentication Successful!</h1>
                    <p>You are now logged in to DashSpace CLI.</p>
                    <p><strong>Welcome back!</strong></p>
                    <div class="close-notice">
                        You can close this window and return to your terminal.
                    </div>
                </div>
                <script>
                    setTimeout(() => {
                        if (window.close) window.close();
                    }, 3000);
                </script>
            </body>
            </html>
        `))

		authResponse := &api.AuthResponse{
			Token: token,
			User: api.User{
				Username: username,
				Email:    email,
			},
		}

		resultChan <- authResponse
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errorChan <- err
		}
	}()

	// Wait for result or timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	select {
	case authResponse := <-resultChan:
		server.Shutdown(context.Background())
		return authResponse, nil
	case err := <-errorChan:
		server.Shutdown(context.Background())
		return nil, err
	case <-ctx.Done():
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("authentication timeout after 5 minutes")
	}
}

func openDeepLink(link string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{link}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", link}
	case "linux":
		cmd = "xdg-open"
		args = []string{link}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}

// Server Side - internal/features/auth/cli/deeplink_handler.go

// Frontend Integration - App Side (React/Vue/etc)

// 1. Deep Link Handler in App
/*
// src/utils/deepLinkHandler.ts
export class DeepLinkHandler {
    static handleAuthCLI(params: URLSearchParams) {
        const state = params.get('state');
        if (!state) {
            console.error('No state parameter in deep link');
            return;
        }

        // Show CLI auth dialog/modal
        this.showCLIAuthDialog(state);
    }

    static async showCLIAuthDialog(state: string) {
        // Show modal asking user to confirm CLI authentication
        const confirmed = await showConfirmDialog({
            title: 'CLI Authentication Request',
            message: 'A CLI tool is requesting access to your account. Do you want to authorize this?',
            confirmText: 'Authorize',
            cancelText: 'Cancel'
        });

        if (confirmed) {
            await this.authorizeCliAccess(state);
        }
    }

    static async authorizeCliAccess(state: string) {
        try {
            const response = await fetch('/api/auth/cli/generate-token', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${getAuthToken()}`
                },
                body: JSON.stringify({ state })
            });

            const data = await response.json();

            if (data.success) {
                // Open the callback URL to complete CLI authentication
                window.open(data.callback_url, '_blank');

                showSuccessMessage('CLI access authorized successfully!');
            }
        } catch (error) {
            console.error('CLI auth error:', error);
            showErrorMessage('Failed to authorize CLI access');
        }
    }
}

// 2. App Router Integration
// src/App.tsx or main router
useEffect(() => {
    const handleDeepLink = (url: string) => {
        const urlObj = new URL(url);

        if (urlObj.protocol === 'dashspace:' && urlObj.pathname === '/auth/cli') {
            DeepLinkHandler.handleAuthCLI(urlObj.searchParams);
        }
    };

    // Listen for deep links
    if (window.electronAPI) {
        window.electronAPI.onDeepLink(handleDeepLink);
    }
}, []);

// 3. Electron Main Process Integration
// src/main/index.ts
import { app, protocol } from 'electron';

app.setAsDefaultProtocolClient('dashspace');

app.on('open-url', (event, url) => {
    event.preventDefault();
    // Send deep link to renderer
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send('deep-link', url);
    }
});

// Windows/Linux handling
app.on('second-instance', (event, commandLine) => {
    const url = commandLine.find(arg => arg.startsWith('dashspace://'));
    if (url && mainWindow) {
        mainWindow.webContents.send('deep-link', url);
        mainWindow.show();
    }
});
*/
