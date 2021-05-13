/*
Copyright 2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package keypaths defines several keypaths used by multiple Teleport services.
package keypaths

import (
	"path/filepath"
	"strings"
)

// keypath constants aren't exported in order to force
// helper function usage and maintain consistency.
const (
	// sessionKeyDir is a sub-directory where session keys are stored
	sessionKeyDir = "keys"
	// sDirSuffix is the suffix of a sub-directory where SSH certificates are stored.
	sshDirSuffix = "-ssh"
	// fileNameKnownHosts is a file where known hosts are stored.
	fileNameKnownHosts = "known_hosts"
	// fileExtTLSCert is the suffix/extension of a file where a TLS cert is stored.
	fileExtTLSCert = "-x509.pem"
	// fileNameTLSCerts is a file where TLS Cert Authorities are stored.
	fileNameTLSCerts = "certs.pem"
	// fileExtCert is the suffix/extension of a file where an SSH Cert is stored.
	fileExtSSHCert = "-cert.pub"
	// fileExtPub is the extension of a file where a public key is stored.
	fileExtPub = ".pub"
	// appDirSuffix is the suffix of a sub-directory where app TLS certs are stored.
	appDirSuffix = "-app"
	// db DirSuffix is the suffix of a sub-directory where db  TLS certs are stored.
	dbDirSuffix = "-db"
	// kubeDirSuffix is the suffix of a sub-directory where kube TLS certs are stored.
	kubeDirSuffix = "-kube"
)

// Here's the file layout of all these keypaths.
// ~/.tsh/							 --> default base directory
// ├── known_hosts                   --> trusted certificate authorities (their keys) in a format similar to known_hosts
// └── keys
//    ├── one.example.com            --> Proxy hostname
//    │   ├── certs.pem              --> TLS CA certs for the Teleport CA
//    │   ├── foo                    --> RSA Private Key for user "foo"
//    │   ├── foo.pub                --> Public Key
//    │   ├── foo-x509.pem           --> TLS client certificate for Auth Server
//    │   ├── foo-ssh                --> SSH certs for user "foo"
//    │   │   ├── root-cert.pub      --> SSH cert for Teleport cluster "root"
//    │   │   └── leaf-cert.pub      --> SSH cert for Teleport cluster "leaf"
//    │   ├── foo-kube               --> Kubernetes certs for user "foo"
//    │   │   ├── root               --> Kubernetes certs for Teleport cluster "root"
//    │   │   │   ├── kubeA-x509.pem --> TLS cert for Kubernetes cluster "kubeA"
//    │   │   │   └── kubeB-x509.pem --> TLS cert for Kubernetes cluster "kubeB"
//    │   │   └── leaf               --> Kubernetes certs for Teleport cluster "leaf"
//    │   │       └── kubeC-x509.pem --> TLS cert for Kubernetes cluster "kubeC"
//    │   └── foo-db                 --> Database access certs for user "foo"
//    │       ├── root               --> Database access certs for cluster "root"
//    │       │   ├── dbA-x509.pem   --> TLS cert for database service "dbA"
//    │       │   └── dbB-x509.pem   --> TLS cert for database service "dbB"
//    │       └── leaf               --> Database access certs for cluster "leaf"
//    │           └── dbC-x509.pem   --> TLS cert for database service "dbC"
//    └── two.example.com
//        ├── certs.pem
//        ├── bar
//        ├── bar.pub
//        ├── bar-x509.pem
//        └── bar-ssh
//            └── clusterA-cert.pub

// KeyDir returns the path to the keys directory.
//
// <baseDir>/keys
func KeyDir(baseDir string) string {
	return filepath.Join(
		baseDir,
		sessionKeyDir,
	)
}

// KnownHostsPath returns the path to the known hosts file.
//
// <baseDir>/known_hosts
func KnownHostsPath(baseDir string) string {
	return filepath.Join(baseDir, fileNameKnownHosts)
}

// ProxyKeyDir returns the path to the proxy's keys directory.
//
// <baseDir>/keys/<proxy>
func ProxyKeyDir(baseDir, proxy string) string {
	return filepath.Join(
		KeyDir(baseDir),
		proxy,
	)
}

// KeyPath returns the path to the users's private key
// for the given proxy.
//
// <baseDir>/keys/<proxy>/<username>.
func UserKeyPath(baseDir, proxy, username string) string {
	return filepath.Join(
		ProxyKeyDir(baseDir, proxy),
		username,
	)
}

// TLSCertPath returns the path to the users's TLS certificate
// for the given proxy.
//
// <baseDir>/keys/<proxy>/<username>-x509.pem
func TLSCertPath(baseDir, proxy, username string) string {
	return filepath.Join(
		ProxyKeyDir(baseDir, proxy),
		username+fileExtTLSCert,
	)
}

// SSHCAsPath returns the path to the users's SSH CA's certificates
// for the given proxy.
//
// <baseDir>/keys/<proxy>/<username>.pub
func SSHCAsPath(baseDir, proxy, username string) string {
	return filepath.Join(
		ProxyKeyDir(baseDir, proxy),
		username+fileExtPub,
	)
}

// CACertPath returns the path to the users's TLS CA's certificates
// for the given proxy.
//
// <baseDir>/keys/<proxy>/certs.pem
func TLSCAsPath(baseDir, proxy string) string {
	return filepath.Join(ProxyKeyDir(baseDir, proxy), fileNameTLSCerts)
}

// SSHDir returns the path to the user's SSH directory for the given proxy.
//
// <baseDir>/keys/<proxy>/<username>-ssh
func SSHDir(baseDir, proxy, username string) string {
	return filepath.Join(
		ProxyKeyDir(baseDir, proxy),
		username+sshDirSuffix,
	)
}

// SSHCertPath returns the path to the users's SSH certificate
// for the given proxy and cluster.
//
// <baseDir>/keys/<proxy>/<username>-ssh/<cluster>-cert.pub
func SSHCertPath(baseDir, proxy, username, cluster string) string {
	return filepath.Join(
		SSHDir(baseDir, proxy, username),
		cluster+fileExtSSHCert,
	)
}

// AppDir returns the path to the user's app directory
// for the given proxy.
//
// <baseDir>/keys/<proxy>/<username>-app
func AppDir(baseDir, proxy, username string) string {
	return filepath.Join(
		ProxyKeyDir(baseDir, proxy),
		username+appDirSuffix,
	)
}

// AppCertDir returns the path to the user's app cert directory
// for the given proxy and cluster.
//
// <baseDir>/keys/<proxy>/<username>-app/<cluster>
func AppCertDir(baseDir, proxy, username, cluster string) string {
	return filepath.Join(
		ProxyKeyDir(baseDir, proxy),
		cluster,
	)
}

// AppCertPath returns the path to the user's TLS certificate
// for the given proxy, cluster, and app.
//
// <baseDir>/keys/<proxy>/<username>-app/<cluster>/<appname>-x509.pem
func AppCertPath(baseDir, proxy, username, cluster, appname string) string {
	return filepath.Join(
		AppCertDir(baseDir, proxy, username, cluster),
		appname+fileExtTLSCert,
	)
}

// DatabaseDir returns the path to the user's kube directory
// for the given proxy.
//
// <baseDir>/keys/<proxy>/<username>-db
func DatabaseDir(baseDir, proxy, username string) string {
	return filepath.Join(
		ProxyKeyDir(baseDir, proxy),
		username+dbDirSuffix,
	)
}

// DatabaseCertDir returns the path to the user's kube cert directory
// for the given proxy and cluster.
//
// <baseDir>/keys/<proxy>/<username>-db/<cluster>
func DatabaseCertDir(baseDir, proxy, username, cluster string) string {
	return filepath.Join(
		DatabaseDir(baseDir, proxy, username),
		cluster,
	)
}

// DatabaseCertPath returns the path to the user's TLS certificate
// for the given proxy, cluster, and kube cluster.
//
// <baseDir>/keys/<proxy>/<username>-db/<cluster>/<dbname>-x509.pem
func DatabaseCertPath(baseDir, proxy, username, cluster, dbname string) string {
	return filepath.Join(
		DatabaseCertDir(baseDir, proxy, username, cluster),
		dbname+fileExtTLSCert,
	)
}

// KubeDir returns the path to the user's kube directory
// for the given proxy.
//
// <baseDir>/keys/<proxy>/<username>-kube
func KubeDir(baseDir, proxy, username string) string {
	return filepath.Join(
		ProxyKeyDir(baseDir, proxy),
		username+kubeDirSuffix,
	)
}

// KubeCertDir returns the path to the user's kube cert directory
// for the given proxy and cluster.
//
// <baseDir>/keys/<proxy>/<username>-kube/<cluster>
func KubeCertDir(baseDir, proxy, username, cluster string) string {
	return filepath.Join(
		KubeDir(baseDir, proxy, username),
		cluster,
	)
}

// KubeCertPath returns the path to the user's TLS certificate
// for the given proxy, cluster, and kube cluster.
//
// <baseDir>/keys/<proxy>/<username>-kube/<cluster>/<kubename>-x509.pem
func KubeCertPath(baseDir, proxy, username, cluster, kubename string) string {
	return filepath.Join(
		KubeCertDir(baseDir, proxy, username, cluster),
		kubename+fileExtTLSCert,
	)
}

// IdentitySSHCertPath returns the path to the identity file's SSH certificate.
//
// <identity-file-dir>/<path>-cert.pub
func IdentitySSHCertPath(path string) string {
	return path + fileExtSSHCert
}

// TrimPathSuffix trims the suffix/extension off of the given cert path.
func TrimCertPathSuffix(path string) string {
	path = strings.TrimSuffix(path, fileExtTLSCert)
	path = strings.TrimSuffix(path, fileExtSSHCert)
	return path
}
