bin_name := "tg"
install_dir := env_var('HOME') / ".local/bin"

# List available commands
default:
    @just --list

# Build the binary into ./tg
build:
    go build -o {{bin_name}} .

# Build and install to ~/.local/bin
install: build
    mkdir -p {{install_dir}}
    install -m 755 {{bin_name}} {{install_dir}}/{{bin_name}}
    @echo "installed to {{install_dir}}/{{bin_name}}"

# Format, vet, and build — run before committing
check:
    gofmt -l .
    go vet ./...
    go build ./...

# Format all source files in place
fmt:
    gofmt -w .

# Run go test ./... (no tests yet, but wired up for when there are)
test:
    go test ./...

# Remove build artifacts
clean:
    rm -f {{bin_name}}

# Build and run with the given args, e.g. `just run -- config show`
run *args: build
    ./{{bin_name}} {{args}}

# Show current config (masked)
config-show: build
    ./{{bin_name}} config show

# Open the config file in $EDITOR
config-edit:
    ${EDITOR:-vi} ~/.config/{{bin_name}}/config.toml
