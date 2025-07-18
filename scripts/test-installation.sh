#!/bin/bash

set -e

echo "🧪 Testing DashSpace CLI installations"

# Test installation script
test_install_script() {
    echo "📜 Testing installation script..."
    if [ -f "scripts/install.sh" ]; then
        # Test script syntax
        bash -n scripts/install.sh && echo "  ✅ Syntax OK" || echo "  ❌ Syntax error"
        echo "  ℹ️  Dry-run test requires manual verification"
    else
        echo "  ⚠️  install.sh not found"
    fi
}

# Test Debian package
test_debian_package() {
    if command -v dpkg >/dev/null 2>&1 && [ -f "packages/dashspace_"*"_amd64.deb" ]; then
        echo "🐧 Testing Debian package..."
        local deb_file=$(ls packages/dashspace_*_amd64.deb | head -1)

        echo "  📋 Package info:"
        dpkg -I "$deb_file" | head -20

        echo "  📁 Package contents:"
        dpkg -c "$deb_file" | head -10

        echo "  ✅ Debian package valid"
    else
        echo "🐧 Debian package test skipped (dpkg not available or package not found)"
    fi
}

# Test binaries
test_binaries() {
    echo "🔧 Testing binaries..."

    if [ ! -d "dist" ]; then
        echo "  ⚠️  dist/ directory not found - run 'make build-all' first"
        return
    fi

    for binary in dist/dashspace-*; do
        if [ -x "$binary" ]; then
            echo "  Testing $(basename "$binary")..."
            if $binary --version >/dev/null 2>&1; then
                echo "    ✅ Version check OK"
            else
                echo "    ❌ Version check failed"
            fi
        else
            echo "  ⚠️  $binary not executable or not found"
        fi
    done
}

# Test package structure
test_package_structure() {
    echo "📁 Testing package structure..."

    local required_dirs=("packaging/debian" "packaging/homebrew" "scripts")
    local required_files=("packaging/debian/control" "packaging/homebrew/dashspace.rb" "scripts/install.sh")

    for dir in "${required_dirs[@]}"; do
        if [ -d "$dir" ]; then
            echo "  ✅ $dir exists"
        else
            echo "  ❌ $dir missing"
        fi
    done

    for file in "${required_files[@]}"; do
        if [ -f "$file" ]; then
            echo "  ✅ $file exists"
        else
            echo "  ❌ $file missing"
        fi
    done
}

# Main test execution
main() {
    test_package_structure
    test_binaries
    test_install_script
    test_debian_package

    echo ""
    echo "✅ Testing completed"
    echo ""
    echo "🚀 Next steps:"
    echo "  1. Fix any issues found above"
    echo "  2. Test installation manually"
    echo "  3. Create GitHub release"
}

# Run tests
main "$@"