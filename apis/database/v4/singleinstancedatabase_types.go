/*
** Copyright (c) 2023 Oracle and/or its affiliates.
**
** The Universal Permissive License (UPL), Version 1.0
**
** Subject to the condition set forth below, permission is hereby granted to any
** person obtaining a copy of this software, associated documentation and/or data
** (collectively the "Software"), free of charge and under any and all copyright
** rights in the Software, and any and all patent rights owned or freely
** licensable by each licensor hereunder covering either (i) the unmodified
** Software as contributed to or provided by such licensor, or (ii) the Larger
** Works (as defined below), to deal in both
**
** (a) the Software, and
** (b) any piece of software and/or hardware listed in the lrgrwrks.txt file if
** one is included with the Software (each a "Larger Work" to which the Software
** is contributed by such licensors),
**
** without restriction, including without limitation the rights to copy, create
** derivative works of, display, perform, and distribute the Software and make,
** use, sell, offer for sale, import, export, have made, and have sold the
** Software and the Larger Work(s), and to sublicense the foregoing rights on
** either these or other terms.
**
** This license is subject to the following condition:
** The above copyright notice and either this complete permission notice or at
** a minimum a reference to the UPL must be included in all copies or
** substantial portions of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
** IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
** FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
** AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
** LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
** OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
 */

package v4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SingleInstanceDatabaseSpec defines the desired state of SingleInstanceDatabase
type SingleInstanceDatabaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Enum=standard;enterprise;express;free
	Edition string `json:"edition,omitempty"`

	// SID must be alphanumeric (no special characters, only a-z, A-Z, 0-9), and no longer than 12 characters.
	// +k8s:openapi-gen=true
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9]+$`
	// +kubebuilder:validation:MaxLength:=12
	Sid                   string            `json:"sid,omitempty"`
	Charset               string            `json:"charset,omitempty"`
	Pdbname               string            `json:"pdbName,omitempty"`
	LoadBalancer          bool              `json:"loadBalancer,omitempty"`
	ListenerPort          int               `json:"listenerPort,omitempty"`
	TcpsListenerPort      int               `json:"tcpsListenerPort,omitempty"`
	ServiceAnnotations    map[string]string `json:"serviceAnnotations,omitempty"`
	FlashBack             *bool             `json:"flashBack,omitempty"`
	ArchiveLog            *bool             `json:"archiveLog,omitempty"`
	ForceLogging          *bool             `json:"forceLog,omitempty"`
	EnableTCPS            bool              `json:"enableTCPS,omitempty"`
	TcpsCertRenewInterval string            `json:"tcpsCertRenewInterval,omitempty"`
	TcpsTlsSecret         string            `json:"tcpsTlsSecret,omitempty"`

	PrimaryDatabaseRef string `json:"primaryDatabaseRef,omitempty"`
	// +kubebuilder:validation:Enum=primary;standby;clone;truecache
	CreateAs             string   `json:"createAs,omitempty"`
	ReadinessCheckPeriod int      `json:"readinessCheckPeriod,omitempty"`
	ServiceAccountName   string   `json:"serviceAccountName,omitempty"`
	TrueCacheServices    []string `json:"trueCacheServices,omitempty"`

	// +k8s:openapi-gen=true
	Replicas int `json:"replicas,omitempty"`

	NodeSelector  map[string]string                   `json:"nodeSelector,omitempty"`
	AdminPassword SingleInstanceDatabaseAdminPassword `json:"adminPassword,omitempty"`
	Image         SingleInstanceDatabaseImage         `json:"image"`
	Persistence   SingleInstanceDatabasePersistence   `json:"persistence,omitempty"`
	InitParams    *SingleInstanceDatabaseInitParams   `json:"initParams,omitempty"`
	Resources     SingleInstanceDatabaseResources     `json:"resources,omitempty"`

	ConvertToSnapshotStandby bool `json:"convertToSnapshotStandby,omitempty"`
}

type SingleInstanceDatabaseResource struct {
	Cpu    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type SingleInstanceDatabaseResources struct {
	Requests *SingleInstanceDatabaseResource `json:"requests,omitempty"`
	Limits   *SingleInstanceDatabaseResource `json:"limits,omitempty"`
}

// SingleInstanceDatabasePersistence defines the storage size and class for PVC
type SingleInstanceDatabasePersistence struct {
	Size         string `json:"size,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
	// +kubebuilder:validation:Enum=ReadWriteOnce;ReadWriteMany
	AccessMode            string `json:"accessMode,omitempty"`
	DatafilesVolumeName   string `json:"datafilesVolumeName,omitempty"`
	ScriptsVolumeName     string `json:"scriptsVolumeName,omitempty"`
	VolumeClaimAnnotation string `json:"volumeClaimAnnotation,omitempty"`
	SetWritePermissions   *bool  `json:"setWritePermissions,omitempty"`
}

// SingleInstanceDatabaseInitParams defines the Init Parameters
type SingleInstanceDatabaseInitParams struct {
	SgaTarget          int `json:"sgaTarget,omitempty"`
	PgaAggregateTarget int `json:"pgaAggregateTarget,omitempty"`
	CpuCount           int `json:"cpuCount,omitempty"`
	Processes          int `json:"processes,omitempty"`
}

// SingleInstanceDatabaseImage defines the Image source and pullSecrets for POD
type SingleInstanceDatabaseImage struct {
	Version     string `json:"version,omitempty"`
	PullFrom    string `json:"pullFrom"`
	PullSecrets string `json:"pullSecrets,omitempty"`
	PrebuiltDB  bool   `json:"prebuiltDB,omitempty"`
}

// SingleInsatnceAdminPassword defines the secret containing Admin Password mapped to secretKey for Database
type SingleInstanceDatabaseAdminPassword struct {
	SecretName string `json:"secretName"`
	// +kubebuilder:default:="oracle_pwd"
	SecretKey  string `json:"secretKey,omitempty"`
	KeepSecret *bool  `json:"keepSecret,omitempty"`
}

// SingleInstanceDatabaseStatus defines the observed state of SingleInstanceDatabase
type SingleInstanceDatabaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Nodes         []string `json:"nodes,omitempty"`
	Role          string   `json:"role,omitempty"`
	Status        string   `json:"status,omitempty"`
	Replicas      int      `json:"replicas,omitempty"`
	ReleaseUpdate string   `json:"releaseUpdate,omitempty"`
	DgBroker      *string  `json:"dgBroker,omitempty"`
	// +kubebuilder:default:="false"
	DatafilesPatched     string            `json:"datafilesPatched,omitempty"`
	ConnectString        string            `json:"connectString,omitempty"`
	ClusterConnectString string            `json:"clusterConnectString,omitempty"`
	TcpsConnectString    string            `json:"tcpsConnectString,omitempty"`
	StandbyDatabases     map[string]string `json:"standbyDatabases,omitempty"`
	// +kubebuilder:default:="false"
	DatafilesCreated     string `json:"datafilesCreated,omitempty"`
	Sid                  string `json:"sid,omitempty"`
	Edition              string `json:"edition,omitempty"`
	Charset              string `json:"charset,omitempty"`
	Pdbname              string `json:"pdbName,omitempty"`
	InitSgaSize          int    `json:"initSgaSize,omitempty"`
	InitPgaSize          int    `json:"initPgaSize,omitempty"`
	CreatedAs            string `json:"createdAs,omitempty"`
	FlashBack            string `json:"flashBack,omitempty"`
	ArchiveLog           string `json:"archiveLog,omitempty"`
	ForceLogging         string `json:"forceLog,omitempty"`
	OemExpressUrl        string `json:"oemExpressUrl,omitempty"`
	OrdsReference        string `json:"ordsReference,omitempty"`
	PdbConnectString     string `json:"pdbConnectString,omitempty"`
	TcpsPdbConnectString string `json:"tcpsPdbConnectString,omitempty"`
	ApexInstalled        bool   `json:"apexInstalled,omitempty"`
	PrebuiltDB           bool   `json:"prebuiltDB,omitempty"`
	// +kubebuilder:default:=false
	IsTcpsEnabled         bool   `json:"isTcpsEnabled"`
	CertCreationTimestamp string `json:"certCreationTimestamp,omitempty"`
	CertRenewInterval     string `json:"certRenewInterval,omitempty"`
	ClientWalletLoc       string `json:"clientWalletLoc,omitempty"`
	PrimaryDatabase       string `json:"primaryDatabase,omitempty"`
	// +kubebuilder:default:=""
	TcpsTlsSecret string `json:"tcpsTlsSecret"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	InitParams  SingleInstanceDatabaseInitParams  `json:"initParams,omitempty"`
	Persistence SingleInstanceDatabasePersistence `json:"persistence"`

	ConvertToSnapshotStandby bool `json:"convertToSnapshotStandby,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:JSONPath=".status.edition",name="Edition",type="string"
// +kubebuilder:printcolumn:JSONPath=".status.sid",name="Sid",type="string",priority=1
// +kubebuilder:printcolumn:JSONPath=".status.status",name="Status",type="string"
// +kubebuilder:printcolumn:JSONPath=".status.role",name="Role",type="string"
// +kubebuilder:printcolumn:JSONPath=".status.releaseUpdate",name="Version",type="string"
// +kubebuilder:printcolumn:JSONPath=".status.connectString",name="Connect Str",type="string"
// +kubebuilder:printcolumn:JSONPath=".status.pdbConnectString",name="Pdb Connect Str",type="string",priority=1
// +kubebuilder:printcolumn:JSONPath=".status.tcpsConnectString",name="TCPS Connect Str",type="string"
// +kubebuilder:printcolumn:JSONPath=".status.tcpsPdbConnectString",name="TCPS Pdb Connect Str",type="string", priority=1
// +kubebuilder:printcolumn:JSONPath=".status.oemExpressUrl",name="Oem Express Url",type="string"

// SingleInstanceDatabase is the Schema for the singleinstancedatabases API
// +kubebuilder:storageversion
type SingleInstanceDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SingleInstanceDatabaseSpec   `json:"spec,omitempty"`
	Status SingleInstanceDatabaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SingleInstanceDatabaseList contains a list of SingleInstanceDatabase
type SingleInstanceDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SingleInstanceDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SingleInstanceDatabase{}, &SingleInstanceDatabaseList{})
}
