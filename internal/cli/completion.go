package cli

import (
	"fmt"
	"os"
)

// CompletionCmd generates shell completion scripts for gdrv.
type CompletionCmd struct {
	Shell string `arg:"" help:"Shell type (bash, zsh, fish, powershell)" enum:"bash,zsh,fish,powershell"`
}

// Run writes the completion script for the requested shell to stdout.
func (cmd *CompletionCmd) Run(_ *Globals) error {
	switch cmd.Shell {
	case "bash":
		_, err := fmt.Fprint(os.Stdout, bashCompletionScript)
		return err
	case "zsh":
		_, err := fmt.Fprint(os.Stdout, zshCompletionScript)
		return err
	case "fish":
		_, err := fmt.Fprint(os.Stdout, fishCompletionScript)
		return err
	case "powershell":
		_, err := fmt.Fprint(os.Stdout, powershellCompletionScript)
		return err
	default:
		return fmt.Errorf("unsupported shell: %s", cmd.Shell)
	}
}

const bashCompletionScript = `#!/usr/bin/env bash
# gdrv bash completion script
# Source this file or add to ~/.bashrc:
#   source <(gdrv completion bash)

_gdrv() {
    local cur prev words cword
    _init_completion 2>/dev/null || {
        COMPREPLY=()
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
    }

    local top_cmds="files folders auth permissions drives sheets docs slides admin changes labels activity chat sync config about version completion"

    case "${COMP_CWORD}" in
    1)
        COMPREPLY=($(compgen -W "${top_cmds}" -- "${cur}"))
        return 0
        ;;
    2)
        case "${COMP_WORDS[1]}" in
        files)
            COMPREPLY=($(compgen -W "list get upload download delete copy move trash restore revisions search" -- "${cur}"))
            ;;
        folders)
            COMPREPLY=($(compgen -W "list create delete move" -- "${cur}"))
            ;;
        auth)
            COMPREPLY=($(compgen -W "login device service-account status profiles logout" -- "${cur}"))
            ;;
        permissions)
            COMPREPLY=($(compgen -W "list create update delete public audit analyze report bulk search" -- "${cur}"))
            ;;
        drives)
            COMPREPLY=($(compgen -W "list get" -- "${cur}"))
            ;;
        sheets)
            COMPREPLY=($(compgen -W "list get create batch-update values" -- "${cur}"))
            ;;
        docs)
            COMPREPLY=($(compgen -W "list get create read update" -- "${cur}"))
            ;;
        slides)
            COMPREPLY=($(compgen -W "list get create read update replace" -- "${cur}"))
            ;;
        admin)
            COMPREPLY=($(compgen -W "users groups" -- "${cur}"))
            ;;
        changes)
            COMPREPLY=($(compgen -W "list start-page-token watch stop" -- "${cur}"))
            ;;
        labels)
            COMPREPLY=($(compgen -W "list get create publish disable file" -- "${cur}"))
            ;;
        activity)
            COMPREPLY=($(compgen -W "query" -- "${cur}"))
            ;;
        chat)
            COMPREPLY=($(compgen -W "spaces messages members" -- "${cur}"))
            ;;
        config)
            COMPREPLY=($(compgen -W "show set reset" -- "${cur}"))
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
            ;;
        esac
        return 0
        ;;
    esac

    COMPREPLY=($(compgen -W "${top_cmds}" -- "${cur}"))
    return 0
}

complete -F _gdrv gdrv
`

const zshCompletionScript = `#compdef gdrv
# gdrv zsh completion script
# Add to ~/.zshrc:
#   source <(gdrv completion zsh)

_gdrv() {
    local -a commands

    if (( CURRENT == 2 )); then
        commands=(
            'files:File operations'
            'folders:Folder operations'
            'auth:Authentication commands'
            'permissions:Permission operations'
            'drives:Manage Shared Drives'
            'sheets:Google Sheets operations'
            'docs:Google Docs operations'
            'slides:Google Slides operations'
            'admin:Google Workspace Admin SDK operations'
            'changes:Drive Changes API operations'
            'labels:Drive Labels API operations'
            'activity:Drive Activity API operations'
            'chat:Google Chat operations'
            'sync:Sync local folders with Drive'
            'config:Configuration management'
            'about:Display Drive account information and API capabilities'
            'version:Print the version number'
            'completion:Generate shell completion scripts'
        )
        _describe 'command' commands
        return
    fi

    case "${words[2]}" in
    files)
        local -a sub=(list get upload download delete copy move trash restore revisions search)
        _describe 'files subcommand' sub
        ;;
    folders)
        local -a sub=(list create delete move)
        _describe 'folders subcommand' sub
        ;;
    auth)
        local -a sub=(login device service-account status profiles logout)
        _describe 'auth subcommand' sub
        ;;
    permissions)
        local -a sub=(list create update delete public audit analyze report bulk search)
        _describe 'permissions subcommand' sub
        ;;
    drives)
        local -a sub=(list get)
        _describe 'drives subcommand' sub
        ;;
    sheets)
        local -a sub=(list get create batch-update values)
        _describe 'sheets subcommand' sub
        ;;
    docs)
        local -a sub=(list get create read update)
        _describe 'docs subcommand' sub
        ;;
    slides)
        local -a sub=(list get create read update replace)
        _describe 'slides subcommand' sub
        ;;
    admin)
        local -a sub=(users groups)
        _describe 'admin subcommand' sub
        ;;
    changes)
        local -a sub=(list start-page-token watch stop)
        _describe 'changes subcommand' sub
        ;;
    labels)
        local -a sub=(list get create publish disable file)
        _describe 'labels subcommand' sub
        ;;
    activity)
        local -a sub=(query)
        _describe 'activity subcommand' sub
        ;;
    chat)
        local -a sub=(spaces messages members)
        _describe 'chat subcommand' sub
        ;;
    config)
        local -a sub=(show set reset)
        _describe 'config subcommand' sub
        ;;
    completion)
        local -a sub=(bash zsh fish)
        _describe 'shell' sub
        ;;
    esac
}

_gdrv "$@"
`

const fishCompletionScript = `# gdrv fish completion script
# Add to ~/.config/fish/config.fish:
#   gdrv completion fish | source

# Disable file completions for gdrv
complete -c gdrv -f

# Top-level commands
complete -c gdrv -n "__fish_use_subcommand" -a files       -d "File operations"
complete -c gdrv -n "__fish_use_subcommand" -a folders     -d "Folder operations"
complete -c gdrv -n "__fish_use_subcommand" -a auth        -d "Authentication commands"
complete -c gdrv -n "__fish_use_subcommand" -a permissions -d "Permission operations"
complete -c gdrv -n "__fish_use_subcommand" -a drives      -d "Manage Shared Drives"
complete -c gdrv -n "__fish_use_subcommand" -a sheets      -d "Google Sheets operations"
complete -c gdrv -n "__fish_use_subcommand" -a docs        -d "Google Docs operations"
complete -c gdrv -n "__fish_use_subcommand" -a slides      -d "Google Slides operations"
complete -c gdrv -n "__fish_use_subcommand" -a admin       -d "Google Workspace Admin SDK operations"
complete -c gdrv -n "__fish_use_subcommand" -a changes     -d "Drive Changes API operations"
complete -c gdrv -n "__fish_use_subcommand" -a labels      -d "Drive Labels API operations"
complete -c gdrv -n "__fish_use_subcommand" -a activity    -d "Drive Activity API operations"
complete -c gdrv -n "__fish_use_subcommand" -a chat        -d "Google Chat operations"
complete -c gdrv -n "__fish_use_subcommand" -a sync        -d "Sync local folders with Drive"
complete -c gdrv -n "__fish_use_subcommand" -a config      -d "Configuration management"
complete -c gdrv -n "__fish_use_subcommand" -a about       -d "Display Drive account information and API capabilities"
complete -c gdrv -n "__fish_use_subcommand" -a version     -d "Print the version number"
complete -c gdrv -n "__fish_use_subcommand" -a completion  -d "Generate shell completion scripts"

# files subcommands
complete -c gdrv -n "__fish_seen_subcommand_from files" -a list      -d "List files"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a get       -d "Get file metadata"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a upload    -d "Upload a file"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a download  -d "Download a file"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a delete    -d "Delete a file"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a copy      -d "Copy a file"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a move      -d "Move a file"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a trash     -d "Move file to trash"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a restore   -d "Restore file from trash"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a revisions -d "List file revisions"
complete -c gdrv -n "__fish_seen_subcommand_from files" -a search    -d "Search files"

# auth subcommands
complete -c gdrv -n "__fish_seen_subcommand_from auth" -a login           -d "Login with OAuth2"
complete -c gdrv -n "__fish_seen_subcommand_from auth" -a device          -d "Login with device code flow"
complete -c gdrv -n "__fish_seen_subcommand_from auth" -a service-account -d "Authenticate with service account"
complete -c gdrv -n "__fish_seen_subcommand_from auth" -a status          -d "Show authentication status"
complete -c gdrv -n "__fish_seen_subcommand_from auth" -a profiles        -d "Manage authentication profiles"
complete -c gdrv -n "__fish_seen_subcommand_from auth" -a logout          -d "Logout and clear credentials"

# completion subcommands (shells)
complete -c gdrv -n "__fish_seen_subcommand_from completion" -a bash -d "Generate bash completion script"
complete -c gdrv -n "__fish_seen_subcommand_from completion" -a zsh  -d "Generate zsh completion script"
complete -c gdrv -n "__fish_seen_subcommand_from completion" -a fish -d "Generate fish completion script"

# Global flags
complete -c gdrv -l profile   -d "Authentication profile to use"
complete -c gdrv -l drive-id  -d "Shared Drive ID to operate in"
complete -c gdrv -l output    -d "Output format (json, table)" -a "json table"
complete -c gdrv -s q -l quiet   -d "Suppress non-essential output"
complete -c gdrv -s v -l verbose -d "Enable verbose logging"
complete -c gdrv -l debug     -d "Enable debug output"
complete -c gdrv -l dry-run   -d "Show what would be done without making changes"
complete -c gdrv -s f -l force   -d "Force operation without confirmation"
complete -c gdrv -s y -l yes     -d "Answer yes to all prompts"
complete -c gdrv -l json      -d "Output in JSON format"
`

const powershellCompletionScript = `# gdrv PowerShell completion script
# Add to your PowerShell profile:
#   gdrv completion powershell | Out-String | Invoke-Expression
#
# To install permanently:
#   gdrv completion powershell > $PROFILE.CurrentUserCurrentHost

function _gdrv_completer {
    param($wordToComplete, $commandAst, $cursorPosition)

    $commands = @(
        'files', 'folders', 'auth', 'permissions', 'drives', 'sheets', 'docs', 'slides',
        'admin', 'groups', 'changes', 'labels', 'activity', 'chat', 'forms', 'people',
        'gmail', 'calendar', 'tasks', 'appscript', 'ai', 'meet', 'logging', 'monitoring',
        'iamadmin', 'sync', 'config', 'about', 'version', 'completion'
    )

    $parts = $commandAst.CommandElements | ForEach-Object { $_.ToString() }
    $parts = $parts[1..($parts.Length-1)]  # Remove 'gdrv' itself

    # First level completion
    if ($parts.Count -eq 0 -or ($parts.Count -eq 1 -and -not $wordToComplete.Contains(' '))) {
        $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'Command', $_)
        }
        return
    }

    # Second level completion
    $subcommands = switch ($parts[0]) {
        'files' { @('list', 'list-trashed', 'get', 'upload', 'download', 'delete', 'copy', 'move', 'trash', 'restore', 'revisions', 'search', 'export', 'touch') }
        'folders' { @('list', 'create', 'delete', 'move', 'empty-tree') }
        'auth' { @('login', 'device', 'service-account', 'status', 'profiles', 'logout') }
        'permissions' { @('list', 'create', 'update', 'delete', 'public', 'audit', 'analyze', 'report', 'bulk', 'search') }
        'drives' { @('list', 'get') }
        'sheets' { @('list', 'list-by-label', 'get', 'create', 'batch-update', 'values') }
        'docs' { @('list', 'get', 'create', 'read', 'update') }
        'slides' { @('list', 'get', 'create', 'read', 'update', 'replace') }
        'admin' { @('users', 'groups') }
        'groups' { @('list', 'get', 'create', 'update', 'delete', 'list-members', 'add-member', 'remove-member') }
        'changes' { @('list', 'start-page-token', 'watch', 'stop') }
        'labels' { @('list', 'get', 'create', 'publish', 'disable', 'file') }
        'activity' { @('query') }
        'chat' { @('spaces', 'messages', 'members') }
        'forms' { @('list', 'get', 'create') }
        'people' { @('list', 'get', 'create', 'delete') }
        'gmail' { @('list', 'get', 'send', 'draft') }
        'calendar' { @('list', 'get', 'create', 'update', 'delete', 'events') }
        'tasks' { @('lists', 'tasklists') }
        'appscript' { @('list', 'get', 'create', 'deploy', 'run') }
        'ai' { @('generate', 'chat') }
        'meet' { @('spaces', 'conferences') }
        'logging' { @('logs', 'sinks') }
        'monitoring' { @('metrics', 'alerts', 'dashboards') }
        'iamadmin' { @('roles', 'permissions') }
        'config' { @('show', 'set', 'reset') }
        'completion' { @('bash', 'zsh', 'fish', 'powershell') }
        default { @() }
    }

    if ($parts.Count -eq 1 -or ($parts.Count -eq 2 -and -not $wordToComplete.Contains(' ') -and $parts[1] -eq $wordToComplete)) {
        $subcommands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'Command', $_)
        }
        return
    }
}

Register-ArgumentCompleter -Native -CommandName gdrv -ScriptBlock _gdrv_completer
`
