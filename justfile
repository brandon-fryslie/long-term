build:
    go build -o long-term .

install: build
    mkdir -p ~/.local/bin
    cp long-term ~/.local/bin/

release TYPE:
    #!/usr/bin/env bash
    set -euo pipefail

    # Get the latest tag, or use v0.0.0 if no tags exist
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

    # Remove the 'v' prefix for version manipulation
    VERSION=${LATEST_TAG#v}

    # Split version into components
    IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

    # Increment based on TYPE
    case "{{TYPE}}" in
        major)
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            ;;
        minor)
            MINOR=$((MINOR + 1))
            PATCH=0
            ;;
        patch)
            PATCH=$((PATCH + 1))
            ;;
        *)
            echo "Error: TYPE must be major, minor, or patch"
            exit 1
            ;;
    esac

    NEW_TAG="v${MAJOR}.${MINOR}.${PATCH}"

    echo "Current version: $LATEST_TAG"
    echo "New version: $NEW_TAG"
    echo ""
    read -p "Create and push tag $NEW_TAG? [y/N] " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag "$NEW_TAG"
        git push origin "$NEW_TAG"
        echo "✓ Tagged and pushed $NEW_TAG"
        echo "GitHub Actions will build the release shortly"
    else
        echo "Cancelled"
        exit 1
    fi
