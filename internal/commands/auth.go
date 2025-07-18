package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/devlyspace/devly-cli/internal/api"
	"github.com/devlyspace/devly-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Se connecter à votre compte DashSpace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleLogin()
		},
	}
}

func NewLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Se déconnecter",
		RunE: func(cmd *cobra.Command, args []string) error {
			config.ClearAuth()
			fmt.Println("✅ Déconnexion réussie")
			return nil
		},
	}
}

func NewWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Afficher l'utilisateur connecté",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.GetConfig()
			if cfg.AuthToken == "" {
				fmt.Println("❌ Non connecté. Utilisez 'dashspace login'")
				return nil
			}

			client := api.NewClient()
			user, err := client.GetCurrentUser()
			if err != nil {
				fmt.Printf("❌ Erreur: %v\n", err)
				return nil
			}

			fmt.Printf("👤 Connecté en tant que: %s\n", user.Username)
			fmt.Printf("📧 Email: %s\n", user.Email)
			return nil
		},
	}
}

func handleLogin() error {
	fmt.Println("🔐 Connexion à DashSpace")
	fmt.Println()

	// Demander email
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	// Demander mot de passe
	fmt.Print("Mot de passe: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("erreur lecture mot de passe: %v", err)
	}
	password := string(passwordBytes)
	fmt.Println()

	client := api.NewClient()
	authResponse, err := client.Login(email, password)
	if err != nil {
		return fmt.Errorf("échec connexion: %v", err)
	}

	cfg := config.GetConfig()
	cfg.AuthToken = authResponse.Token
	cfg.Username = authResponse.User.Username
	cfg.Email = authResponse.User.Email

	if err := config.SaveConfig(); err != nil {
		return fmt.Errorf("erreur sauvegarde config: %v", err)
	}

	fmt.Printf("✅ Connecté en tant que %s\n", authResponse.User.Username)
	return nil
}
