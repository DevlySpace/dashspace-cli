package commands

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/devlyspace/dashspace-cli/internal/api"
	"github.com/devlyspace/dashspace-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ====== COMMAND DEFINITIONS ======

func NewLoginCmd() *cobra.Command {
	var webLogin bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to DashSpace",
		Long:  "Authenticate with DashSpace to publish modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			if webLogin {
				return handleWebLogin()
			}
			return handleLogin()
		},
	}

	cmd.Flags().BoolVarP(&webLogin, "web", "w", false, "Use web browser for authentication")

	return cmd
}

func NewLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout from DashSpace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleLogout()
		},
	}
}

func NewWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Display current logged in user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleWhoami()
		},
	}
}

// ====== HANDLER FUNCTIONS ======

func handleLogin() error {
	fmt.Println("🔐 Login to DashSpace")
	fmt.Println("───────────────────")

	reader := bufio.NewReader(os.Stdin)

	// Get email
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	// Get password
	fmt.Print("Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("error reading password: %v", err)
	}
	password := strings.TrimSpace(string(bytePassword))
	fmt.Println() // New line after password

	// Authenticate
	fmt.Println("\n⏳ Authenticating...")
	client := api.NewClient()
	authResponse, err := client.Login(email, password)
	if err != nil {
		return fmt.Errorf("❌ Authentication failed: %v", err)
	}

	// Save credentials
	cfg := config.GetConfig()
	cfg.AuthToken = authResponse.Token
	cfg.Username = authResponse.User.Username
	cfg.Email = authResponse.User.Email

	if err := config.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	fmt.Printf("\n✅ Successfully logged in as %s\n", cfg.Username)
	fmt.Println("📦 You can now publish modules to DashSpace!")

	return nil
}

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

func handleLogout() error {
	cfg := config.GetConfig()

	if cfg.AuthToken == "" {
		fmt.Println("❌ You are not logged in")
		return nil
	}

	// Clear auth info
	config.ClearAuth()

	fmt.Println("✅ Successfully logged out")
	return nil
}

func handleWhoami() error {
	cfg := config.GetConfig()

	if cfg.AuthToken == "" {
		fmt.Println("❌ Not logged in")
		fmt.Println("💡 Run 'dashspace login' to authenticate")
		return nil
	}

	fmt.Println("👤 Current user:")
	fmt.Printf("   Username: %s\n", cfg.Username)
	fmt.Printf("   Email: %s\n", cfg.Email)

	// Optionally verify token is still valid
	client := api.NewClient()
	if _, err := client.GetProfile(); err != nil {
		fmt.Println("\n⚠️  Your session may have expired. Please login again.")
	} else {
		fmt.Println("\n✅ Session is active")
	}

	return nil
}

// ====== UTILITY FUNCTIONS ======

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

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func saveAuthConfig(authResponse *api.AuthResponse) error {
	cfg := config.GetConfig()

	cfg.AuthToken = authResponse.Token
	cfg.Username = authResponse.User.Username
	cfg.Email = authResponse.User.Email

	if err := config.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %v", err)
	}

	fmt.Printf("\n✅ Successfully logged in as %s\n", cfg.Username)
	return nil
}
