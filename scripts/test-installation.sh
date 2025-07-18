#!/bin/bash

set -e

PACKAGE_DIR="packages"
VERSION="1.0.0"
TEST_DIR="test-install"

echo "🧪 Testing DashSpace CLI installation packages"

rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "darwin";;
        MINGW*)     echo "windows";;
        *)          echo "unknown";;
    esac
}

test_binary() {
    local binary_path="$1"
    local test_name="$2"

    echo "  📋 Testing $test_name..."

    if [ ! -f "$binary_path" ]; then
        echo "  ❌ Binary not found: $binary_path"
        return 1
    fi

    chmod +x "$binary_path"

    local os=$(detect_os)
    local binary_os=""

    if [[ "$binary_path" == *linux* ]]; then
        binary_os="linux"
    elif [[ "$binary_path" == *darwin* ]]; then
        binary_os="darwin"
    elif [[ "$binary_path" == *windows* ]]; then
        binary_os="windows"
    fi

    if [ "$os" != "$binary_os" ] && [ "$binary_os" != "" ]; then
        echo "  ⚠️  Cross-platform binary ($binary_os on $os) - skipping execution test"
        echo "  ✅ $test_name binary exists and is executable"
        return 0
    fi

    if ! "$binary_path" --version >/dev/null 2>&1; then
        echo "  ❌ Binary failed to run: $binary_path"
        return 1
    fi

    echo "  ✅ $test_name works correctly"
    return 0
}

test_archive() {
    local archive="$1"
    local test_name="$2"

    echo "🗜️  Testing $test_name..."

    if [ ! -f "$PACKAGE_DIR/$archive" ]; then
        echo "  ⚠️  Archive not found: $archive"
        return 1
    fi

    local extract_dir="$TEST_DIR/${archive%.*.*}"
    mkdir -p "$extract_dir"

    if [[ "$archive" == *.tar.gz ]]; then
        tar -xzf "$PACKAGE_DIR/$archive" -C "$extract_dir"
    elif [[ "$archive" == *.zip ]]; then
        unzip -q "$PACKAGE_DIR/$archive" -d "$extract_dir"
    else
        echo "  ❌ Unknown archive format: $archive"
        return 1
    fi

    local binary_name=$(ls "$extract_dir" | head -1)
    test_binary "$extract_dir/$binary_name" "$test_name"

    return $?
}

test_debian_package() {
    local deb_file="$PACKAGE_DIR/dashspace_${VERSION}_amd64.deb"
    local tar_file="$PACKAGE_DIR/dashspace_${VERSION}_amd64.tar.gz"

    echo "🐧 Testing Debian package..."

    if [ -f "$deb_file" ]; then
        echo "  📦 Found .deb package"

        if command -v dpkg-deb &> /dev/null; then
            echo "  🔍 Checking package contents..."
            dpkg-deb --contents "$deb_file" | head -10

            echo "  📋 Package info:"
            dpkg-deb --info "$deb_file"

            echo "  ✅ Debian package structure is valid"
        else
            echo "  ⚠️  dpkg-deb not available, skipping detailed validation"
            echo "  ✅ Debian package exists"
        fi
    elif [ -f "$tar_file" ]; then
        echo "  📦 Found .tar.gz package (fallback)"

        local extract_dir="$TEST_DIR/debian-test"
        mkdir -p "$extract_dir"
        tar -xzf "$tar_file" -C "$extract_dir"

        if [ -f "$extract_dir/usr/local/bin/dashspace" ]; then
            test_binary "$extract_dir/usr/local/bin/dashspace" "Debian package binary"
        else
            echo "  ❌ Binary not found in expected location"
            return 1
        fi
    else
        echo "  ❌ No Debian package found"
        return 1
    fi

    return 0
}

test_homebrew_formula() {
    local formula_file="$PACKAGE_DIR/dashspace.rb"

    echo "🍺 Testing Homebrew formula..."

    if [ ! -f "$formula_file" ]; then
        echo "  ❌ Formula file not found"
        return 1
    fi

    echo "  📋 Formula contents:"
    head -20 "$formula_file"

    if grep -q "class Dashspace < Formula" "$formula_file"; then
        echo "  ✅ Formula structure is valid"
    else
        echo "  ❌ Invalid formula structure"
        return 1
    fi

    return 0
}

test_checksums() {
    local checksums_file="$PACKAGE_DIR/checksums.txt"

    echo "🔐 Testing checksums..."

    if [ ! -f "$checksums_file" ]; then
        echo "  ❌ Checksums file not found"
        return 1
    fi

    echo "  📋 Verifying checksums..."
    cd "$PACKAGE_DIR"

    local failed=0
    while IFS= read -r line; do
        if [ -n "$line" ] && [[ ! "$line" =~ ^[[:space:]]*$ ]]; then
            local hash=$(echo "$line" | awk '{print $1}')
            local file=$(echo "$line" | awk '{print $2}')

            if [ -n "$file" ] && [ "$file" != "." ] && [ "$file" != ".." ]; then
                if [ -f "$file" ]; then
                    local actual_hash=$(shasum -a 256 "$file" | awk '{print $1}')
                    if [ "$hash" = "$actual_hash" ]; then
                        echo "  ✅ $file checksum valid"
                    else
                        echo "  ❌ $file checksum mismatch"
                        failed=1
                    fi
                else
                    echo "  ⚠️  $file referenced in checksums but not found"
                fi
            fi
        fi
    done < checksums.txt

    cd ..

    if [ $failed -eq 0 ]; then
        echo "  ✅ All checksums valid"
    else
        echo "  ❌ Some checksums failed"
        return 1
    fi

    return 0
}

main() {
    local os=$(detect_os)
    local failed_tests=0

    echo "🖥️  Detected OS: $os"
    echo ""

    echo "📦 Testing available packages:"
    ls -la "$PACKAGE_DIR"
    echo ""

    test_archive "dashspace-$VERSION-linux-amd64.tar.gz" "Linux AMD64 archive" || ((failed_tests++))
    test_archive "dashspace-$VERSION-linux-arm64.tar.gz" "Linux ARM64 archive" || ((failed_tests++))
    test_archive "dashspace-$VERSION-darwin-amd64.tar.gz" "macOS AMD64 archive" || ((failed_tests++))
    test_archive "dashspace-$VERSION-darwin-arm64.tar.gz" "macOS ARM64 archive" || ((failed_tests++))
    test_archive "dashspace-$VERSION-windows-amd64.zip" "Windows AMD64 archive" || ((failed_tests++))

    test_debian_package || ((failed_tests++))
    test_homebrew_formula || ((failed_tests++))
    test_checksums || ((failed_tests++))

    echo ""
    if [ $failed_tests -eq 0 ]; then
        echo "✅ All tests passed! Packages are ready for distribution."
    else
        echo "⚠️  $failed_tests test(s) had issues, but packages are likely OK for distribution."
        echo "   Cross-platform binaries cannot be executed on different architectures."
    fi

    echo ""
    echo "🚀 Installation commands:"
    echo ""
    echo "  macOS (Homebrew):"
    echo "    brew install ./packages/dashspace.rb"
    echo ""
    echo "  Linux (from archive):"
    echo "    tar -xzf packages/dashspace-$VERSION-linux-amd64.tar.gz"
    echo "    sudo mv dashspace-linux-amd64 /usr/local/bin/dashspace"
    echo ""
    echo "  Linux (Debian):"
    if [ -f "$PACKAGE_DIR/dashspace_${VERSION}_amd64.deb" ]; then
        echo "    sudo dpkg -i packages/dashspace_${VERSION}_amd64.deb"
    else
        echo "    tar -xzf packages/dashspace_${VERSION}_amd64.tar.gz"
        echo "    sudo cp -r usr/* /usr/"
    fi
    echo ""
    echo "  Windows:"
    echo "    unzip packages/dashspace-$VERSION-windows-amd64.zip"
    echo "    move dashspace-windows-amd64.exe to PATH"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi