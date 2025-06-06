package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/spf13/cobra"
)

var (
	printOutput bool
	strict      bool
	debug       bool

	supportedShells = []string{
		"bash",
		"zsh",
		"fish",
		"powershell",
		"pwsh",
		"cmd",
		"nu",
		"elvish",
		"xonsh",
	}

	initCmd = createInitCmd()
)

func init() {
	RootCmd.AddCommand(initCmd)
}

func createInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [bash|zsh|fish|powershell|pwsh|cmd|nu|elvish|xonsh]",
		Short: "Initialize your shell and config",
		Long: `Initialize your shell and config.

See the documentation to initialize your shell: https://ohmyposh.dev/docs/installation/prompt.`,
		ValidArgs: supportedShells,
		Args:      NoArgsOrOneValidArg,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}

			runInit(args[0])
		},
	}

	initCmd.Flags().BoolVarP(&printOutput, "print", "p", false, "print the init script")
	initCmd.Flags().BoolVarP(&strict, "strict", "s", false, "run in strict mode")
	initCmd.Flags().BoolVar(&debug, "debug", false, "enable/disable debug mode")

	_ = initCmd.MarkPersistentFlagRequired("config")

	return initCmd
}

func runInit(sh string) {
	var startTime time.Time

	if debug {
		startTime = time.Now()
		log.Enable()
		log.Debug("debug mode enabled")
	}

	configFile := config.Path(configFlag)
	cfg, hash := config.Load(configFile, sh, false)

	// set session ID here so we can reuse the same logic
	// to initialize the caches everywhere
	sessionID := uuid.NewString()
	os.Setenv("POSH_SESSION_ID", sessionID)

	flags := &runtime.Flags{
		Shell:     sh,
		Config:    configFile,
		Strict:    strict,
		Debug:     debug,
		SaveCache: true,
		Init:      true,
		SessionID: sessionID,
	}

	env := &runtime.Terminal{}
	env.Init(flags)

	template.Init(env, cfg.Var, cfg.Maps)

	defer func() {
		template.SaveCache()
		env.Close()
	}()

	feats := cfg.Features(env)
	flags.ConfigHash = fmt.Sprintf("%s.%s", hash, feats.Hash())

	var output string

	switch {
	case printOutput, debug:
		output = shell.PrintInit(env, feats, &startTime)
	default:
		output = shell.Init(env, feats)
	}

	if silent {
		return
	}

	fmt.Print(output)
}
