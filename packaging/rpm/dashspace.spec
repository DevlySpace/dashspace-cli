Name:           dashspace-cli
Version:        1.0.0
Release:        1%{?dist}
Summary:        Official DashSpace CLI for creating and publishing modules

License:        MIT
URL:            https://dashspace.space
Source0:        dashspace-linux-amd64

BuildArch:      x86_64
Requires:       glibc >= 2.17

%description
DashSpace CLI enables developers to create, test and publish
modules for the DashSpace ecosystem. It includes ready-to-use
templates and integration with the Buildy editor.

%prep
# No preparation needed for binary

%build
# No compilation needed

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}/usr/local/bin
cp %{SOURCE0} %{buildroot}/usr/local/bin/dashspace

%files
/usr/local/bin/dashspace

%post
echo "✅ DashSpace CLI installed successfully"
echo "Get started with: dashspace login"

%changelog
* Mon Jan 15 2024 DashSpace Team <tech@dashspace.space> - 1.0.0-1
- Initial RPM package release