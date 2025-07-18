package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func NewPreviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preview",
		Short: "Ouvrir le module dans Buildy pour preview",
		RunE: func(cmd *cobra.Command, args []string) error {
			return openInBuildy()
		},
	}
}

func openInBuildy() error {
	// Vérifier qu'on est dans un dossier de module
	if _, err := os.Stat("devly.json"); err != nil {
		return fmt.Errorf("❌ Pas de fichier devly.json trouvé. Êtes-vous dans un dossier de module ?")
	}

	// Obtenir le chemin absolu du module
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("erreur obtention du répertoire courant: %v", err)
	}

	moduleName := filepath.Base(pwd)

	buildyURL := fmt.Sprintf("http://localhost:3000/buildy?module=%s&path=%s", moduleName, pwd)

	fmt.Printf("🚀 Ouverture de '%s' dans Buildy...\n", moduleName)
	fmt.Printf("📂 Chemin: %s\n", pwd)
	fmt.Printf("🔗 URL: %s\n", buildyURL)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", buildyURL)
	case "linux":
		cmd = exec.Command("xdg-open", buildyURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", buildyURL)
	default:
		fmt.Printf("⚠️  Système non supporté. Ouvrez manuellement: %s\n", buildyURL)
		return nil
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("⚠️  Impossible d'ouvrir automatiquement. URL: %s\n", buildyURL)
		return nil
	}

	fmt.Println("✅ Buildy ouvert dans votre navigateur")
	return nil
}
