#!/bin/sh
set -eu

repo="${AGENTMETER_INSTALL_REPO:-why19970628/agentmeter}"
install_dir="${AGENTMETER_INSTALL_DIR:-/usr/local/bin}"
requested_version="${AGENTMETER_VERSION:-latest}"

need() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "agentmeter install: missing required command: $1" >&2
		exit 1
	fi
}

detect_os() {
	case "$(uname -s)" in
	Darwin) echo "darwin" ;;
	Linux) echo "linux" ;;
	*)
		echo "agentmeter install: unsupported OS: $(uname -s)" >&2
		exit 1
		;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
	x86_64 | amd64) echo "amd64" ;;
	arm64 | aarch64) echo "arm64" ;;
	*)
		echo "agentmeter install: unsupported architecture: $(uname -m)" >&2
		exit 1
		;;
	esac
}

latest_tag() {
	curl -fsSL "https://api.github.com/repos/${repo}/releases/latest" |
		sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' |
		head -n 1
}

checksum_file() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		shasum -a 256 "$1" | awk '{print $1}'
	fi
}

need curl
need sed
need awk
need tar
if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
	echo "agentmeter install: missing required command: sha256sum or shasum" >&2
	exit 1
fi

os="$(detect_os)"
arch="$(detect_arch)"
tag="$requested_version"
if [ "$tag" = "latest" ]; then
	tag="$(latest_tag)"
fi
if [ -z "$tag" ]; then
	echo "agentmeter install: could not determine release version" >&2
	exit 1
fi

version="${tag#v}"
archive="agentmeter_${version}_${os}_${arch}.tar.gz"
base_url="https://github.com/${repo}/releases/download/${tag}"
tmpdir="$(mktemp -d)"

cleanup() {
	rm -rf "$tmpdir"
}
trap cleanup EXIT INT TERM

echo "Installing agentmeter ${tag} for ${os}/${arch}..." >&2
curl -fsSL "${base_url}/${archive}" -o "${tmpdir}/${archive}"
curl -fsSL "${base_url}/checksums.txt" -o "${tmpdir}/checksums.txt"

expected="$(awk -v name="$archive" '$2 == name {print $1}' "${tmpdir}/checksums.txt")"
if [ -z "$expected" ]; then
	echo "agentmeter install: checksum not found for ${archive}" >&2
	exit 1
fi
actual="$(checksum_file "${tmpdir}/${archive}")"
if [ "$actual" != "$expected" ]; then
	echo "agentmeter install: checksum mismatch for ${archive}" >&2
	exit 1
fi

tar -xzf "${tmpdir}/${archive}" -C "$tmpdir" agentmeter
if mkdir -p "$install_dir" 2>/dev/null && [ -w "$install_dir" ]; then
	mv "${tmpdir}/agentmeter" "${install_dir}/agentmeter"
	chmod 0755 "${install_dir}/agentmeter"
else
	need sudo
	sudo mkdir -p "$install_dir"
	sudo mv "${tmpdir}/agentmeter" "${install_dir}/agentmeter"
	sudo chmod 0755 "${install_dir}/agentmeter"
fi

echo "agentmeter installed to ${install_dir}/agentmeter" >&2
