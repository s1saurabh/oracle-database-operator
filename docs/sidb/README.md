# Managing Oracle Single Instance Databases with Oracle Database Operator for Kubernetes

Oracle Database Operator for Kubernetes (`OraOperator`) includes the Single Instance Database Controller, which enables provisioning, cloning, and patching of Oracle Single Instance Databases on Kubernetes. It also enables configuring the database for Oracle REST Data Services with Oracle APEX development platform. The following sections explain the setup and functionality of the operator

  * [Prerequisites](#prerequisites)
    * [Mandatory Resource Privileges](#mandatory-resource-privileges)
    * [Optional Resource Privileges](#optional-resource-privileges)
    * [OpenShift Security Context Constraints](#openshift-security-context-constraints)
  * [SingleInstanceDatabase Resource](#singleinstancedatabase-resource)
    * [Create a Database](#create-a-database)
      * [New Database](#new-database)
      * [Pre-built Database](#pre-built-database)
      * [XE Database](#xe-database)
      * [Free Database](#free-database)
      * [Free Lite Database](#free-lite-database)
      * [Oracle True Cache](#oracle-true-cache)
    * [Connecting to Database](#connecting-to-database)
    * [Database Persistence (Storage) Configuration Options](#database-persistence-storage-configuration-options)
      * [Dynamic Persistence](#dynamic-persistence)
        * [Storage Expansion](#storage-expansion)
      * [Static Persistence](#static-persistence)
    * [Configuring a Database](#configuring-a-database)
      * [Switching Database Modes](#switching-database-modes)
      * [Changing Init Parameters](#changing-init-parameters)
    * [Clone a Database](#clone-a-database)
    * [Patch a Database](#patch-a-database)
    * [Delete a Database](#delete-a-database)
    * [Advanced Database Configurations](#advanced-database-configurations)
      * [Run Database with Multiple Replicas](#run-database-with-multiple-replicas)
      * [Database Pod Resource Management](#database-pod-resource-management)
      * [Setup Database with LoadBalancer](#setup-database-with-loadbalancer)
      * [Enabling TCPS Connections](#enabling-tcps-connections)
      * [Specifying Custom Ports](#specifying-custom-ports)
      * [Setup Data Guard Configuration for a Single Instance Database](#setup-data-guard-configuration-for-a-single-instance-database)
        * [Create a Standby Database](#create-a-standby-database)
        * [Create a Data Guard Configuration](#create-a-data-guard-configuration)
        * [Perform a Switchover](#perform-a-switchover)
        * [Enable Fast-Start Failover](#enable-fast-start-failover)
        * [Convert Standby to Snapshot Standby](#convert-standby-to-snapshot-standby)
        * [Static Data Guard Connect String](#static-data-guard-connect-string)
        * [Patch Primary and Standby databases](#patch-primary-and-standby-databases)
        * [Delete the Data Guard Configuration](#delete-the-data-guard-configuration)
      * [Execute Custom Scripts](#execute-custom-scripts)
  * [OracleRestDataService Resource](#oraclerestdataservice-resource)
    * [REST Enable a Database](#rest-enable-a-database)
      * [Provision ORDS](#provision-ords)
      * [Database API](#database-api)
      * [MongoDB API](#mongodb-api)
      * [Advanced Usages](#advanced-usages)
        * [Oracle Data Pump](#oracle-data-pump)
        * [REST Enabled SQL](#rest-enabled-sql)
        * [Database Actions](#database-actions)
    * [APEX Installation](#apex-installation)
    * [Delete ORDS](#delete-ords)
  * [Maintenance Operations](#maintenance-operations)
  * [Additional Information](#additional-information)


## Prerequisites

Oracle strongly recommends that you comply with the [prerequisites](./PREREQUISITES.md) and the following requirements

  ### Mandatory Resource Privileges

  Single Instance Database(sidb) controller mandatorily requires the following Kubernetes resource privileges:

  | Resources | Privileges  |
  | --- | --- |
  | Pods | create delete get list patch update watch | 
  | Containers | create delete get list patch update watch |
  | PersistentVolumeClaims | create delete get list patch update watch | 
  | Services | create delete get list patch update watch | 
  | Secrets | create delete get list patch update watch | 
  | Events | create patch |

  For managing the required levels of access, configure [role binding](../../README.md#create-role-bindings-for-access-management)

  ### Optional Resource Privileges

  Single Instance Database(`sidb`) controller optionally requires the following Kubernetes resource privileges, depending on the functionality being used:

  | Functionality | Resources | Privileges |
  | --- | --- | --- | 
  | NodePort Services | Nodes | list watch |
  | Storage Expansion with block volumes | StorageClasses | get list watch |
  | Custom Scripts Execution | PersistentVolumes | get list watch |


  For exposing the database using Nodeport services, apply [RBAC](../../rbac/node-rbac.yaml)
  ```sh
    kubectl apply -f rbac/node-rbac.yaml
  ```
  For automatic storage expansion of block volumes, apply [RBAC](../../rbac/storage-class-rbac.yaml)
  ```sh
    kubectl apply -f rbac/storage-class-rbac.yaml
  ```
  For automatic execution of custom scripts after database setup or startup, apply [RBAC](../../rbac/persistent-volume-rbac.yaml)
  ```sh
    kubectl apply -f rbac/persistent-volume-rbac.yaml
  ```
  
  ### OpenShift Security Context Constraints

  OpenShift requires additional Security Context Constraints (SCC) for deploying and managing the `SingleInstanceDatabase` resource. To create the appropriate SCCs before deploying the `SingleInstanceDatabase` resource, complete these steps:

  1. Create a new project/namespace for deploying the `SingleInstanceDatabase` resource

  ```sh
    oc new-project sidb-ns
  ```

  **Note:** OpenShift recommends that you should not deploy in namespaces starting with `kube`, `openshift` and the `default` namespace.

  2. Apply the file [openshift_rbac.yaml](../../config/samples/sidb/openshift_rbac.yaml) with cluster-admin user privileges.

  ```sh
    oc apply -f openshift-rbac.yaml
  ```

  Running this example procedure results in creation of SCC (Security Context Constraints) and serviceaccount `sidb-sa` in the namespace `sidb-ns`, which has access to the SCC.

  **Note:** This configuration yaml file example binds the SCC to the serviceaccount `sidb-sa` in namespace `sidb-ns`. For any other project/namespace, you must update the file appropriately with the namespace before applying this example.

  3. Set the `serviceAccountName` attribute to `sidb-sa` and the namespace to `sidb-ns` in **[config/samples/sidb/singleinstancedatabase.yaml](../../config/samples/sidb/singleinstancedatabase.yaml)** before deploying the SingleInstanceDatabase resource.

## SingleInstanceDatabase Resource

The Oracle Database Operator creates the `SingleInstanceDatabase` as a custom resource. Doing this enables Oracle Database to be managed as a native Kubernetes object. In this document, we will refer to the `SingleInstanceDatabase` resource as the database.

### Resource Details

#### Database List
To list databases, use the following command as an example, where the database names are `sidb-sample` and `sidb-sample-clone`, which are the names we will use as database names in command examples:

```sh
$ kubectl get singleinstancedatabases -o name

  singleinstancedatabase.database.oracle.com/sidb-sample  
  singleinstancedatabase.database.oracle.com/sidb-sample-clone

```

#### Quick Status
To obtain a quick database status, use the following command as an example:

```sh
$ kubectl get singleinstancedatabase sidb-sample

NAME          EDITION      STATUS    VERSION      CONNECT STR                 TCPS CONNECT STR       OEM EXPRESS URL
sidb-sample   Enterprise   Healthy   19.3.0.0.0   10.0.25.54:1521/ORCL1       Unavailable            https://10.0.25.54:5500/em
```

#### Detailed Status
To obtain a detailed database status, use the following command as an example:

```sh
$ kubectl describe singleinstancedatabase sidb-sample-clone

  Name:         sidb-sample-clone
  Namespace:    default
  Labels:       <none>
  Annotations:  <none>
  API Version:  database.oracle.com/v1alpha1
  Kind:         SingleInstanceDatabase
  Metadata: ....
  Spec: ....
  Status:
      Cluster Connect String:  sidb-sample-clone.default:1521/ORCL1C
      Conditions:
        Last Transition Time:   (YYYY-MM-DD)T(HH:MM:SS)Z
        Message:                Waiting for database to be ready
        Observed Generation:    2
        Reason:                 LastReconcileCycleQueued
        Status:                 True
        Type:                   ReconcileQueued
        Last Transition Time:   2021-06-30T11:07:56Z
        Message:                processing datapatch execution
        Observed Generation:    3
        Reason:                 LastReconcileCycleBlocked
        Status:                 True
        Type:                   ReconcileBlocked
        Last Transition Time:   (YYYY-MM-DD)T(HH:MM:SS)Z
        Message:                no reconcile errors
        Observed Generation:    3
        Reason:                 LastReconcileCycleCompleted
        Status:                 True
        Type:                   ReconcileComplete
      Connect String:          10.0.25.58:1521/ORCL1C
      Datafiles Created:       true
      Datafiles Patched:       true
      Edition:                 Enterprise
      Flash Back:              true
      Force Log:               false
      Oem Express URL:         https://10.0.25.58:5500/em
      Pdb Name:                orclpdb1
      Release Update:          19.11.0.0.0
      Replicas:                2
      Role:                    PRIMARY
      Sid:                     ORCL1C
      Status:                  Healthy
  Events:
      Type     Reason                 Age                    From                    Message
      ----     ------                 ----                   ----                    -------
      Normal   Database Pending       35m (x2 over 35m)      SingleInstanceDatabase  waiting for database pod to be ready
      Normal   Database Creating      27m (x24 over 34m)     SingleInstanceDatabase  waiting for database to be ready
      Normal   Database Ready         22m                    SingleInstanceDatabase  database open on pod sidb-sample-clone-133ol scheduled on node 10.0.10.6
      Normal   Datapatch Pending      21m                    SingleInstanceDatabase  datapatch execution pending
      Normal   Datapatch Executing    20m                    SingleInstanceDatabase  datapatch begin execution
      Normal   Datapatch Done         8s                     SingleInstanceDatabase  datafiles patched from 19.3.0.0.0 to 19.11.0.0.0 : SUCCESS

```

### Template YAML
  
The template `.yaml` file for Single Instance Database (Enterprise and Standard Editions), including all the configurable options, is available at:
**[`config/samples/sidb/singleinstancedatabase.yaml`](./../../config/samples/sidb/singleinstancedatabase.yaml)**

**Note:** 
The `adminPassword` field in the above `singleinstancedatabase.yaml`example file refers to a Secret for the SYS, SYSTEM and PDBADMIN users of the Single Instance Database. This Secret is required when you provision a new database, or when you clone an existing database.

Create this Secret using the following command as an example:

    kubectl create secret generic db-admin-secret --from-literal=oracle_pwd=<specify password here>

This command creates a Secret named `db-admin-secret`, with the key `oracle_pwd` mapped to the actual password specified in the command.

### Create a Database

#### New Database

To provision a new database instance on the Kubernetes cluster, use the example **[`config/samples/sidb/singleinstancedatabase_create.yaml`](../../config/samples/sidb/singleinstancedatabase_create.yaml)**.

1. Log into [Oracle Container Registry](https://container-registry.oracle.com/) and accept the license agreement for the Database image; ignore if you have accepted the license agreement already.

2. If you have not already done so, create an image pull secret for the Oracle Container Registry:

    ```sh
    $ kubectl create secret docker-registry oracle-container-registry-secret --docker-server=container-registry.oracle.com --docker-username='<oracle-sso-email-address>' --docker-password='<container-registry-auth-token>' --docker-email='<oracle-sso-email-address>'
      
      secret/oracle-container-registry-secret created 
    ```
    Note: Generate the auth token from user profile section on top right of the page after logging into container-registry.oracle.com
    
    This secret can also be created from the docker config.json or from podman auth.json after a successful login
    ```sh
    docker login container-registry.oracle.com
    kubectl create secret generic oracle-container-registry-secret  --from-file=.dockerconfigjson=.docker/config.json --type=kubernetes.io/dockerconfigjson
    ```
    or
    ```sh
    podman login container-registry.oracle.com
    kubectl create secret generic oracle-container-registry-secret  --from-file=.dockerconfigjson=${XDG_RUNTIME_DIR}/containers/auth.json --type=kubernetes.io/dockerconfigjson
    ```
3. Provision a new database instance on the cluster by using the following command:

    ```sh
    $ kubectl apply -f singleinstancedatabase_create.yaml

    singleinstancedatabase.database.oracle.com/sidb-sample created
    ```

**Note:** 
- For ease of use, the storage class **oci-bv** is specified in the **[`singleinstancedatabase_create.yaml`](../../config/samples/sidb/singleinstancedatabase_create.yaml)**. This storage class facilitates dynamic provisioning of the OCI block volumes on the Oracle OKE for persistent storage of the database. The supported access mode for this class is `ReadWriteOnce`. For other cloud providers, you can similarly use their dynamic provisioning storage classes.
- It is beneficial to have the database replica pods more than or equal to the number of available nodes if `ReadWriteMany` access mode is used with the OCI NFS volume. By doing so, the pods get distributed on different nodes and the database image is downloaded on all those nodes. This helps in reducing time for the database fail-over if the active database pod dies.
- Supports Oracle Database Enterprise Edition (19.3.0), and later releases.
- To pull the database image faster from the container registry, so that you can bring up the SIDB instance quickly, you can use the `container-registry mirror` of the corresponding cluster's region. For example, if the cluster exists in Mumbai region, then you can use the `container-registry-bom.oracle.com` mirror. For more information on container-registry mirrors, see: [https://blogs.oracle.com/wim/post/oracle-container-registry-mirrors-in-oracle-cloud-infrastructure](https://blogs.oracle.com/wim/post/oracle-container-registry-mirrors-in-oracle-cloud-infrastructure).
- To update the initialization (init) parameters, such as `sgaTarget` and `pgaAggregateTarget`, see the `initParams` section of the [`singleinstancedatabase.yaml`](../../config/samples/sidb/singleinstancedatabase.yaml) file.

#### Pre-built Database

To provision a new pre-built database instance, use the sample **[`config/samples/sidb/singleinstancedatabase_prebuiltdb.yaml](../../config/samples/sidb/singleinstancedatabase_prebuiltdb.yaml)** file. For example:
```sh
$ kubectl apply -f singleinstancedatabase_prebuiltdb.yaml

  singleinstancedatabase.database.oracle.com/prebuiltdb-sample created
```

This pre-built image includes the data files of the database inside the image itself. As a result, the database startup time of the container is reduced, down to a couple of seconds. The pre-built database image can be very useful in continuous integration/continuous delivery (CI/CD) scenarios, in which databases are used for conducting tests or experiments, and the workflow is simple. 

To build the pre-built database image for the Enterprise/Standard edition, follow these instructions: [Pre-built Database (prebuiltdb) Extension](https://github.com/oracle/docker-images/blob/main/OracleDatabase/SingleInstance/extensions/prebuiltdb/README.md).

#### XE Database
To provision a new Oracle Database Express Edition (XE) database, use the sample **[config/samples/sidb/singleinstancedatabase_express.yaml](../../config/samples/sidb/singleinstancedatabase_express.yaml)** file. For example:

      kubectl apply -f singleinstancedatabase_express.yaml

This command pulls the XE image available in [Oracle Container Registry](https://container-registry.oracle.com/).

**Note:**
- Provisioning Oracle Database Express Edition is supported for release 21c (21.3.0) only. Oracle Database Free replaces Oracle Database Express Edition.
- For XE database, only single replica mode (i.e. `replicas: 1`) is supported.
- For XE database, you **cannot change** the init parameters, such as `cpuCount, processes, sgaTarget or pgaAggregateTarget`.

#### Free Database
To provision new Oracle Database Free, use the sample **[config/samples/sidb/singleinstancedatabase_free.yaml](../../config/samples/sidb/singleinstancedatabase_free.yaml)** file. For example:

      kubectl apply -f singleinstancedatabase_free.yaml

This command pulls the Free image available in [Oracle Container Registry](https://container-registry.oracle.com/).

#### Free Lite Database
To provision new Oracle Database Free Lite, use the sample **[config/samples/sidb/singleinstancedatabase_free-lite.yaml](../../config/samples/sidb/singleinstancedatabase_free-lite.yaml)** file. For example:

      kubectl apply -f singleinstancedatabase_free-lite.yaml

This command pulls the Free lite image available in [Oracle Container Registry](https://container-registry.oracle.com/).

**Note:**
- Provisioning Oracle Database Free is supported for release 23.3.0 and later releases.
- For Free database, only single replica mode (such as `replicas: 1`) is supported.
- For Free database, you **cannot change** the init parameters. These include parameters such as `cpuCount, processes, sgaTarget or pgaAggregateTarget`.
- Oracle Enterprise Manager Express (OEM Express) is not supported in release 23.3.0 and later releases. 

#### Oracle True Cache
Oracle True Cache is an in-memory, consistent, and automatically managed cache for Oracle Database.  
To provision a True Cache instance for Oracle Free Database in Kubernetes, use the sample **[`config/samples/sidb/singleinstancedatabase_free-truecache.yaml`](../../config/samples/sidb/singleinstancedatabase_free-truecache.yaml)** file. For example

      kubectl apply -f singleinstancedatabase_free-truecache.yaml

#### Additional Information
You are required to specify the database administrative user (admin) password Secret in the corresponding YAML file. The default values mentioned in the `adminPassword.secretName` fields of [`singleinstancedatabase_create.yaml`](../../config/samples/sidb/singleinstancedatabase_create.yaml), [`singleinstancedatabase_prebuiltdb.yaml`](../../config/samples/sidb/singleinstancedatabase_prebuiltdb.yaml), [`singleinstancedatabase_express.yaml`](../../config/samples/sidb/singleinstancedatabase_express.yaml) and [`singleinstancedatabse_free.yaml`](../../config/samples/sidb/singleinstancedatabase_free.yaml) files are `db-admin-secret`, `prebuiltdb-admin-secret`, `xedb-admin-secret` and `free-admin-secret` respectively. You can create these Secrets manually by using the sample command mentioned in the [`Template YAML`](#template-yaml) section. Alternatively, you can create these Secrets by filling in the passwords in the **[`singleinstancedatabase_secrets.yaml`](../../config/samples/sidb/singleinstancedatabase_secrets.yaml)** file and applying them using the following command:

```bash
kubectl apply -f singleinstancedatabase_secrets.yaml
```

### Connecting to Database

Creating a new database instance takes a while. When the `status` column returns the response `Healthy`, the database is open for connections.

```sh
$ kubectl get singleinstancedatabase sidb-sample -o "jsonpath={.status.status}"
  
  Healthy
```

Clients can obtain the connect string to the CDB from `.status.connectString`, and the connect string to the PDB from `.status.pdbConnectString`. For example:

```sh
$ kubectl get singleinstancedatabase sidb-sample -o "jsonpath={.status.connectString}"

  10.0.25.54:1521/ORCL
```
```sh
$ kubectl get singleinstancedatabase sidb-sample -o "jsonpath={.status.pdbConnectString}"

  10.0.25.54:1521/ORCLPDB
```

To connect to the database using the connect strings returned by the commands above, you can use any supported client, or use SQLPlus. For example:
```sh
$ sqlplus sys/<.spec.adminPassword>@10.0.25.54:1521/ORCL as sysdba

SQL*Plus: Release 19.0.0.0.0 - Production on Wed May 4 16:00:49 2022
Version 19.14.0.0.0

Copyright (c) 1982, 2021, Oracle.  All rights reserved.


Connected to:
Oracle Database 21c Express Edition Release 21.0.0.0.0 - Production
Version 21.3.0.0.0

SQL>
```
**Note:** The `<.spec.adminPassword>` above refers to the database password for SYS, SYSTEM and PDBADMIN users, which in turn represented by `spec` section's `adminPassword` field of the **[config/samples/sidb/singleinstancedatabase.yaml](../config/samples/sidb/../../../../config/samples/sidb/singleinstancedatabase.yaml)** file.

The Oracle Database inside the container also has Oracle Enterprise Manager Express (OEM Express) as a basic observability console. To access OEM Express, start the browser, and paste in a URL similar to the following example:

```sh
$ kubectl get singleinstancedatabase sidb-sample -o "jsonpath={.status.oemExpressUrl}"

  https://10.0.25.54:5500/em
```
**Note:** OEM Express is not available for 23.3.0 and later releases

### Database Persistence (Storage) Configuration Options
You can configure database persistence in the following two ways:
- Dynamic Persistence Provisioning
- Static Persistence Provisioning

#### Dynamic Persistence
In **Dynamic Persistence Provisioning**, a persistent volume is provisioned by mentioning a storage class. For example, **oci-bv** storage class is specified in the **[singleinstancedatabase_create.yaml](../../config/samples/sidb/singleinstancedatabase_create.yaml)** file. This storage class facilitates dynamic provisioning of the OCI block volumes. The supported access mode for this class is `ReadWriteOnce`. For other cloud providers, you can similarly use their dynamic provisioning storage classes.
                     
**Note:** 
- Generally, the `Reclaim Policy` of such dynamically provisioned volumes is `Delete`. These volumes are deleted when their corresponding database deployment is deleted. To retain volumes, use static provisioning, as explained in the Block Volume Static Provisioning section.
- In **Minikube**, the dynamic persistence provisioning class is **standard**.

#### Storage Expansion
When using dynamic persistence, you can at any time scale up your persistent volumes by simply patching the singleinstancedatabase resource using the following command :
```sh
$ kubectl patch singleinstancedatabase sidb-sample -p '{"spec":{"persistence":{"size":"100Gi"}}}' --type=merge
```

**Note:**
- Storage expansion requires the storage class to be configured with `allowVolumeExpansion:true`
- Storage expansion requires read and watch access for storage account as mentioned in [prerequisites](#prerequisites)
- User can only scale up a volume/storage and not scale down

#### Static Persistence
In **Static Persistence Provisioning**, you must create a volume manually, and then use the name of this volume with the `<.spec.persistence.datafilesVolumeName>` field, which corresponds to the `datafilesVolumeName` field of the persistence section in the **[`singleinstancedatabase.yaml`](../../config/samples/sidb/singleinstancedatabase.yaml)**. The `Reclaim Policy` of such volumes can be set to `Retain`. When this policy is set, the volume is not deleted when its corresponding deployment is deleted.
For example in **Minikube**, a persistent volume can be provisioned using the following yaml file example:
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: db-vol
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  hostPath:
    path: /data/oradata
```
The persistent volume name (in this case, `db-vol`) can be mentioned in the `datafilesVolumeName` field of the **[`singleinstancedatabase.yaml`](../../config/samples/sidb/singleinstancedatabase.yaml)**. `storageClass` field is not required in this case, and can be left empty.

Static Persistence Provisioning in Oracle Cloud Infrastructure (OCI) is explained in the following subsections:

##### OCI Block Volume Static Provisioning
With block volume static provisioning, you must manually create a block volume resource from the OCI console, and fetch its `OCID`. To create the persistent volume, you can use the following YAML file:
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: block-vol
spec:
  capacity:
    storage: 1024Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  csi:
    driver: blockvolume.csi.oraclecloud.com
    volumeHandle: <OCID of the block volume>
```

**Note:** OCI block volumes are AD (Availability Domain) specific. Ensure that the database is deployed in the same AD as that of its statically provisioned block volume. In dynamic provisioning, this is done automatically.
To provision the database in a specific AD, uncomment the following line from the **[singleinstancedatabase.yaml](../../config/samples/sidb/singleinstancedatabase.yaml)** file:

```yaml
nodeSelector:
   topology.kubernetes.io/zone: PHX-AD-1
```

##### OCI NFS Volume Static Provisioning
Similar to the block volume static provisioning, you have to manually create a file system resource from the OCI console, and fetch its `OCID, Mount Target IP Address and Export Path`. Mention these values in the following YAML file to create the persistent volume:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-vol
spec:
  capacity:
    storage: 1024Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  csi:
    driver: fss.csi.oraclecloud.com
    volumeHandle: "<OCID of the file system>:<Mount Target IP Address>:/<Export Path>"
```

**Note:** 
- Example volumeHandle in the above config file : 

  `volumeHandle: "ocid1.filesystem.oc1.eu_frankfurt_1.aaaaaqe3bj...eaaa:10.0.10.156:/FileSystem-20220713-1036-02"`

- Whenever a mount target is provisioned in OCI, its `Reported Size (GiB)` values are very large. This is visible on the mount target page when logged in to the OCI console. Some applications will fail to install if the results of a space requirements check show too much available disk space. So in the OCI Console, click the little "Pencil" icon besides the **Reported Size** parameter of the Mount Target to specify, in gigabytes (GiB), the maximum capacity reported by file systems exported through this mount target. This setting does not limit the actual amount of data you can store.

- You must open the required ports to access the NFS volume from the K8S cluster. Add the required ports to the security list of the subnet to which your K8S nodes are connected. For more information, see **[Security Lists File Storage](https://docs.oracle.com/en-us/iaas/Content/File/Tasks/securitylistsfilestorage.htm)** for the details.

### Configuring a Database
The `OraOperator` facilitates you to configure the database. Various database configuration options are explained in the following subsections:

#### Switching Database Modes
The following database modes can be updated after the database is created: 

- `flashBack`
- `archiveLog`
- `forceLog`

To change these modes, change their attribute values, and apply the change by using the 
`kubectl apply` or `kubectl edit/patch` commands. 

**Caution**: Enable `archiveLog` mode before setting `flashBack` to `ON`, and set `flashBack` to `OFF` before disabling `archiveLog` mode.

For example:

```sh
$ kubectl patch singleinstancedatabase sidb-sample --type merge -p '{"spec":{"forceLog": true}}' 

  singleinstancedatabase.database.oracle.com/sidb-sample patched
```
Check the Database Config Status by using the following command:

```sh
$ kubectl get singleinstancedatabase sidb-sample -o "jsonpath=[{.status.archiveLog}, {.status.flashBack}, {.status.forceLog}]"

  [true, true, true]
```

#### Changing Init Parameters

The following database initialization parameters can be updated after the database is created:

- sgaTarget
- pgaAggregateTarget
- cpuCount
- processes. 

Change their attribute values and apply using `kubectl apply` or `kubectl edit/patch` commands.

**Note:**
The value for the initialization parameter `sgaTarget` that you provide should be within the range set by [sga_min_size, sga_max_size]. If the value you provide is not in that range, then `sga_target` is not updated to the value you specify for `sgaTarget`.

#### Immutable YAML Attributes

The following attributes cannot be modified after creating the Single Instance Database instance: 

- `sid`
- `edition`
- `charset`
- `pdbName`
- `primaryDatabaseRef`

If you attempt to change one of these attributes, then you receive an error similar to the following:

```sh
$ kubectl --type=merge -p '{"spec":{"sid":"ORCL1"}}' patch singleinstancedatabase sidb-sample 

  The SingleInstanceDatabase "sidb-sample" is invalid: spec.sid: Forbidden: cannot be changed
```

### Clone a Database

To create copies of your existing database quickly, you can use the cloning functionality. A cloned database is an exact, block-for-block copy of the source database. Cloning is much faster than creating a fresh database and copying over the data.

To quickly clone the existing database `sidb-sample` we previously created for this document, use the sample **[`config/samples/sidb/singleinstancedatabase_clone.yaml`](../../config/samples/sidb/singleinstancedatabase_clone.yaml)** file.

For example:

```sh
  
$ kubectl apply -f singleinstancedatabase_clone.yaml

  singleinstancedatabase.database.oracle.com/sidb-sample-clone created
```

**Note:**
- To clone a database, the source database must have archiveLog mode set to true.
- The clone database can specify a database image that is different from the source database. In such cases, cloning is supported only between databases of the same major release.
- Only enterprise and standard editions support cloning.

### Patch a Database

Databases running in your cluster and managed by the Oracle Database operator can be patched or rolled back between release updates of the same major release. To patch databases, specify an image of the higher release update. To roll back databases, specify an image of the lower release update.

Patched Oracle Docker images can be built by using this [patching extension](https://github.com/oracle/docker-images/tree/main/OracleDatabase/SingleInstance/extensions/patching).

#### Patch

To patch an existing database, edit and apply the **[`config/samples/sidb/singleinstancedatabase_patch.yaml`](../../config/samples/sidb/singleinstancedatabase_patch.yaml)** file of the database resource/object either by specifying a new release update for image attributes, or by running the following command:

```sh
kubectl --type=merge -p '{"spec":{"image":{"pullFrom":"patched-image:tag","pullSecrets":"pull-secret"}}}' patch singleinstancedatabase sidb-sample

singleinstancedatabase.database.oracle.com/sidb-sample patched

```

After patching is complete, the database pods are restarted with the new release update image.

**Note:**
- Only Enterprise and Standard Editions support patching.

#### Patch after Cloning
    
To clone and patch the database at the same time, clone your source database by using the [cloning existing database](#clone-existing-database) method, and specify a new release image for the cloned database. Use this method to ensure there are no patching related issues impacting your database performance or functionality.
    
#### Datapatch Status

Patching/Rollback operations are complete when the datapatch tool completes patching or rollback of the data files. Check the data files patching status
and current release update version using the following commands

```sh
$ kubectl get singleinstancedatabase sidb-sample -o "jsonpath={.status.datafilesPatched}"

  true
  
$ kubectl get singleinstancedatabase sidb-sample -o "jsonpath={.status.releaseUpdate}"

  19.3.0.0.0
```

#### Rollback
You can roll back to a prior database version by specifying the old image in the `image` field of the **[`config/samples/sidb/singleinstancedatabase_patch.yaml`](../../config/samples/sidb/singleinstancedatabase_patch.yaml)** file, and applying it using the following command:

```bash
kubectl apply -f singleinstancedatabase_patch.yaml
```

This can also be done using the following command:

```sh
kubectl --type=merge -p '{"spec":{"image":{"pullFrom":"old-image:tag","pullSecrets":"pull-secret"}}}' patch singleinstancedatabase sidb-sample

singleinstancedatabase.database.oracle.com/sidb-sample patched

```

### Delete a Database
To delete the database, run the following command :

```bash
kubectl delete singleinstancedatabase.database.oracle.com sidb-sample
```
This command will delete the database pods and associated service.

### Advanced Database Configurations
Some advanced database configuration scenarios are as follows:

#### Run Database with Multiple Replicas
In multiple replicas mode, more than one pod is created for the database. Setting the replica count equal to or more than the number of worker nodes helps in distributing the replicas accross all the nodes that have access to the database persistent storage volume.  
The database is open and mounted by one of the replica pods. Other replica pods have the database instance started but not mounted, and serve to provide a quick cold fail-over in case the active pod goes down.

To enable multiple replicas, update the replica attribute in the `.yaml`, and apply by using the `kubectl apply` or `kubectl scale` commands.

The following table depicts the fail over matrix for any destructive operation to the primary replica pod

| Pod Destructive Operation | Pod Restart/FailOver|
  | --- | --- | 
  | Database instance crash | Yes | 
  | Force delete pod with zero grace period | Yes | 
  | Gracefully delete pod | Yes |
  | Node running primary replica dies | Yes | 
  | Direct shutdown [All modes] | Yes | 
  | Maintenance shutdown [All modes] | No |
  | PDB close | No |
  
**Note:** 
- Maintence shutdown/startup can be run by using the scripts `/home/oracle/shutDown.sh` and `/home/oracle/startUp.sh` 
- This functionality requires the [k8s extension](https://github.com/oracle/docker-images/tree/main/OracleDatabase/SingleInstance/extensions/k8s) extended images. The database image from the container registry `container-registry.oracle.com` includes the K8s extension.
- Because Oracle Database Express Edition (XE) does not support [k8s extension](https://github.com/oracle/docker-images/tree/main/OracleDatabase/SingleInstance/extensions/k8s), it does not support multiple replicas.
- If the `ReadWriteOnce` access mode is used, then all the replicas will be scheduled on the same node where the persistent volume would be mounted.
- If the `ReadWriteMany` access mode is used, then all the replicas will be distributed on different nodes. For this reason, Oracle recommends that you have replicas more than or equal to the number of the nodes, because the database image is downloaded on all those nodes. This is beneficial in quick cold fail-over scenario (when the active pod dies) as the image would already be available on that node.

#### Database Pod Resource Management
When creating a Single Instance Database, you can specify the CPU and memory resources needed by the database pod. These specified resources are passed to the `kube-scheduler` so that the pod is scheduled on one of the pods that has the required resources available. To use database pod resource management, specify values for the `resources` attributes in the [`config/samples/sidb/singleinstancedatabase.yaml`](../../config/samples/sidb/singleinstancedatabase.yaml) file, and apply it.

#### Database Pod Resource Management
When creating a Single Instance Database you can specify the cpu and memory resources needed by the database pod. These specified resources are passed to the `kube-scheduler` so that the pod gets scheduled on one of the pods that has the required resources available. To use database pod resource management specify values for the `resources` attributes in the [config/samples/sidb/singleinstancedatabase.yaml](../../config/samples/sidb/singleinstancedatabase.yaml) file, and apply it.

#### Setup Database with LoadBalancer
For the Single Instance Database, the default service is the `NodePort` service. You can enable the `LoadBalancer` service by using the `kubectl patch` command.

For example:

```sh
$ kubectl --type=merge -p '{"spec":{"loadBalancer": true}}' patch singleinstancedatabase sidb-sample 

  singleinstancedatabase.database.oracle.com/sidb-sample patched
```

### Enabling TCPS Connections
You can enable TCPS connections in the database by setting the `enableTCPS` field to `true` in the [`config/samples/sidb/singleinstancedatabase.yaml`](../../config/samples/sidb/singleinstancedatabase.yaml) file, and applying it.

Alternatively, you can use the following command:
```bash
kubectl patch --type=merge singleinstancedatabases.database.oracle.com sidb-sample -p '{"spec": {"enableTCPS": true}}'
```
By default self signed certs are used for TCPS connections. The TCPS connections status can also be queried by the following command:
```bash
kubectl get singleinstancedatabase sidb-sample -o "jsonpath={.status.isTcpsEnabled}"
true
```

**With Self Signed Certs**
- When TCPS is enabled, a self-signed certificate is generated and stored in wallets. For users' convenience, a client-side wallet is generated in location `/opt/oracle/oradata/clientWallet/$ORACLE_SID` in the pod.
- The self-signed certificate used with TCPS has validity for 1 year. After the certificate is expired, it will be renewed by the `OraOperator` automatically. Download the wallet again after auto-renewal.
- You can set the certificate renew interval with the help of `tcpsCertRenewInterval` field in the **[config/samples/sidb/singleinstancedatabase.yaml](../../config/samples/sidb/singleinstancedatabase.yaml)** file. The minimum accepted value is 24h, and the maximum value is 8760h (1 year). The certificates used with TCPS will automatically be renewed after this interval. If this field is omitted/commented in the yaml file, the certificates will not be renewed automatically.
- When the certificate gets created/renewed, the `.status.certCreationTimestamp` status variable gets updated accordingly. You can see this timestamp by using the following command:
  ```bash
  kubectl get singleinstancedatabase sidb-sample  -o "jsonpath={.status.certCreationTimestamp}"
  ```

**With User Provided Certs**
- Users can provide custom certs to be used for TCPS connections instead of self signed ones.
- Specify the certs by creating a Kubernetes tls secret resource using following command:
  ```bash
  kubectl create secret tls my-tls-secret --cert=path/to/cert/tls.crt --key=path/to/key/tls.key
  ```
- `tls.crt` is a certificate chain in the order of client, followed by intermediate and then root certificate and `tls.key` is client key.
- Specify the secret created above (`my-tls-secret`) as the value for the attribute `tcpsTlsSecret` in the [config/samples/sidb/singleinstancedatabase_tcps.yaml](../../config/samples/sidb/singleinstancedatabase_tcps.yaml) file, and apply it.

**Connecting to the Database using TCPS**
- Download the wallet from the Persistent Volume (PV) attached with the database pod. The location of the wallet inside the pod is as `/opt/oracle/oradata/clientWallet/$ORACLE_SID`. Let us assume the `ORACLE_SID` is `ORCL1`, and singleinstance database resource name is `sidb-sample` for the upcoming example command. You can copy the wallet to the destination directory by the following command:
  ```bash
  kubectl cp $(kubectl get pods -l app=sidb-sample -o=jsonpath='{.items[0].metadata.name}'):/opt/oracle/oradata/clientWallet/ORCL1 <Wallet Destination directory>
  ```
- This wallet includes the sample `tnsnames.ora` and `sqlnet.ora` files. All the TNS entries for the database (corresponding to the CDB and PDB) reside in the `tnsnames.ora` file. Switch to the downloaded wallet directory and set the `TNS_ADMIN` environment variable to point to the current directory as follows:
  ```bash
  cd <Wallet Destination directory>
  export TNS_ADMIN=$(pwd)
  ```
  After this, connect with SQL*Plus, using the following example commands:
  ```bash
  sqlplus sys@ORCL1 as sysdba
  ```

### Specifying Custom Ports
As mentioned in the section [Setup Database with LoadBalancer](#setup-database-with-loadbalancer), there are two kubernetes services possible for the database: NodePort and LoadBalancer. You can specify which port to use with these services by editing the `listenerPort` and `tcpsListenerPort` fields of the [`config/samples/sidb/singleinstancedatabase.yaml`](../../config/samples/sidb/singleinstancedatabase.yaml) file.

`listenerPort` is intended for normal database connections. Similarly, `tcpsListenerPort` is intended for TCPS database connections.

If the `LoadBalancer` is enabled, then the `listenerPort`, and `tcpsListenerPort` will be the opened ports on the Load Balancer for normal and TCPS database connections respectively. When the `LoadBalancer` is enabled, the default values of `listenerPort` and `tcpsListenerPort` are 1521 and 2484. 

If the `NodePort` service is enabled, then the `listenerPort`, and `tcpsListenerPort` will be the opened ports on the Kubernetes nodes for for normal and TCPS database connections respectively. In this case, the allowed range for the `listenerPort`, and `tcpsListenerPort` is 30000-32767.

**Note:**
- `listenerPort` and `tcpsListenerPort` cannot have same values.
- `tcpsListenerPort` will come into effect only when TCPS connections are enabled (specifically, the `enableTCPS` field is set in [`config/samples/sidb/singleinstancedatabase.yaml`](../../config/samples/sidb/singleinstancedatabase.yaml) file).
- If TCPS connections are enabled, and `listenerPort` is commented or removed in the [`config/samples/sidb/singleinstancedatabase.yaml`](../../config/samples/sidb/singleinstancedatabase.yaml) file, then only the TCPS endpoint will be exposed.
- If LoadBalancer is enabled, and either `listenerPort` or `tcpsListenerPort` is changed, then it takes some time to complete the work requests (drain existing backend sets and create new ones). During this time, the database connectivity is broken, although `SingleInstanceDatabase` and `LoadBalancer` remain in a healthy state. To check the progress of the work requests, you can by log in to the Cloud provider's console and check the corresponding LoadBalancer.

### Setup Data Guard Configuration for a Single Instance Database

### Create a Standby Database

#### Prerequisites
- Before creating a Standby, ensure that ArchiveLog, FlashBack, and ForceLog on the primary Single Instance Database(`.spec.primaryDatabaseRef`) are turned on.
- Standby database is not supported for TCPS-enabled Primary databases.

#### Template YAML
To create a standby database, edit and apply the example YAML file [`config/samples/sidb/singleinstancedatabase_standby.yaml`](../../config/samples/sidb/singleinstancedatabase_standby.yaml).

**Note:**
- The `adminPassword` field of the above [`config/samples/sidb/singleinstancedatabase_standby.yaml`](../../config/samples/sidb/singleinstancedatabase_standby.yaml) contains an admin password Secret of the primary database referred to for Standby Database creation. By default `keepSecret` is set to `true`, which means that the secret is saved. However, if you want to delete the Secret after the database pod becomes ready, then this Secret will be deleted if the `keepSecret` attribute of `adminPassword` field is set to `false`. .
- Specify the primary database with which the standby database is associateed in the `.spec.primaryDatabaseRef` yaml file.
- The `.spec.createAs` field of the yaml file should be set to "standby".
- Database configuration, such as `Archivelog`, `FlashBack`, `ForceLog`, `TCPS connections`, are not supported for standby database.

#### List Standby Databases
To list the standby databases, use the `get singleinstancedatabase` command.  For example:

```sh
kubectl get singleinstancedatabase

NAME      EDITION      STATUS    ROLE               VERSION      CONNECT STR                  TCPS CONNECT STR           OEM EXPRESS URL
sidb-19   Enterprise   Healthy   PRIMARY            19.3.0.0.0   10.25.0.26:1521/ORCL1        Unavailable                https://10.25.0.26:5500/em
stdby-1   Enterprise   Healthy   PHYSICAL_STANDBY   19.3.0.0.0   10.25.0.27:32392/ORCLS1      Unavailable                https://10.25.0.27:30329/em

```

### Query Primary Database Reference
You can query the corresponding primary database for every standby database. For example:

```sh
kubectl get singleinstancedatabase stdby-1 -o "jsonpath={.status.primaryDatabase}"
sidb-19
```

#### Creation Status
  
 Creating a new standby database instance takes a while. When the 'status' status returns the response "Healthy", the database is open for connections. For example: 

  ```sh
$ kubectl get singleinstancedatabase stdby-1 -o "jsonpath={.status.status}"
   
  Healthy
```

### Create a Data Guard Configuration

#### Template YAML

After creating standbys, set up an Oracle Data Guard (Data Guard) configuration with protection mode, and switch over capability using the following example YAML:
[`config/samples/sidb/dataguardbroker.yaml`](./../../config/samples/sidb/dataguardbroker.yaml)

#### Create DataGuardBroker Resource

To use the Data Guard broker, provision a new `dataguardbroker` custom resource for a single instance database(`.spec.primaryDatabaseRef`) by specifying the appropriate values for the primary and standby databases in the example `.yaml` file, and running the following command:

```sh
$ kubectl create -f dataguardbroker.yaml

  dataguardbroker.database.oracle.com/dataguardbroker-sample created
```
**Note:** The following attributes cannot be patched after you create the `dataguardbroker` resource: `primaryDatabaseRef,    protectionMode`

#### DataguardBroker List

To list the Data Guard broker resources, use the following command:

```sh
  $ kubectl get dataguardbroker -o name

    dataguardbroker.database.oracle.com/dataguardbroker-sample

```

#### Quick Status
You can obtain a quick status of Data Guard broker by using the following command: 

```sh
  $ kubectl get dataguardbroker dataguardbroker-sample

    NAME                      PRIMARY   STANDBYS               PROTECTION MODE      CONNECT STR                       STATUS
    dataguardbroker-sample    ORCL      ORCLS1,ORCLS2          MaxAvailability      10.0.25.85:31555/DATAGUARD        Healthy

```

#### Detailed Status
To obtain more detailed Data Guard broker status, use this command: 

```sh
  $ kubectl describe dataguardbroker dataguardbroker-sample

    Name:         dataguardbroker-sample
    Namespace:    default
    Labels:       <none>
    Annotations:  <none>
    API Version:  database.oracle.com/v1alpha1
    Kind:         DataguardBroker
    Metadata:
      Creation Timestamp:  2023-01-23T04:29:04Z
      Finalizers:
        database.oracle.com/dataguardbrokerfinalizer
      Generation:  3
      Managed Fields:
        API Version:  database.oracle.com/v1alpha1
        Fields Type:  FieldsV1
        fieldsV1:
          ...
        Manager:      manager
        Operation:    Update
        Time:         2023-01-23T04:30:20Z
        API Version:  database.oracle.com/v1alpha1
        Fields Type:  FieldsV1
        fieldsV1:
          ...
        Manager:         kubectl-client-side-apply
        Operation:       Update
        Time:            2023-01-23T04:44:40Z
      Resource Version:  75178376
      UID:               c04a3d88-2018-4f7f-b232-b74d6c3d9479
    Spec:
      Admin Password:
        Keep Secret:  true
        Secret Key:   oracle_pwd
        Secret Name:  db-secret
      Fast Start Failover:      false
      Primary Database Ref:     sidb-sample
      Protection Mode:          MaxAvailability
      Set As Primary Database:  
      Standby Database Refs:
        standby-sample-1
        standby-sample-2
    Status:
      Cluster Connect String:   dataguardbroker-sample.default:1521/DATAGUARD
      External Connect String:  10.0.25.85:31167/DATAGUARD
      Fast Start Failover:      false
      Primary Database:         OR19E3
      Standby Databases:        OR19E3S1,OR19E3S2
      Status:                   Healthy
    Events:
      Type    Reason                       Age                 From             Message
      ----    ------                       ----                ----             -------
      Normal  SUCCESS                      42m                 DataguardBroker  
      Normal  DG Configuration up to date  24m (x13 over 56m)  DataguardBroker  
```
  
### Perform a Switchover

Specify the approppriate database system identifier (SID)  (the SID of one of `.spec.primaryDatabaseRef` , `.spec.standbyDatabaseRefs[]`) to be set primary in the `.spec.setAsPrimaryDatabase` of [`dataguardbroker.yaml`](./../../config/samples/sidb/dataguardbroker.yaml) and apply the yaml file.

When you apply the YAML file, the database you specify will be set to primary. However, if the database specified with the `apply` command is already the primary, then this command has no effect: 

```sh
$ kubectl apply -f dataguardbroker.yaml

  dataguardbroker.database.oracle.com/dataguardbroker-sample apply

```
You can also use the patch command 

```sh
$ kubectl --type=merge -p '{"spec":{"setAsPrimaryDatabase":"ORCLS1"}}' patch dataguardbroker dataguardbroker-sample

  dataguardbroker.database.oracle.com/dataguardbroker-sample patched
```

### Enable Fast-Start Failover

Oracle Data Guard Fast-Start Failover (FSFO) monitors your Oracle Data Guard environments and initiates an automatic failover in the case of an outage.
To enable FSFO, ensure the primary database is in the primary role, set the attribute `.spec.fastStartFailover` to `true` in [`datguardbroker.yaml`](./../../config/samples/sidb/dataguardbroker.yaml), and then apply it. For example: 

```sh
$ kubectl apply -f dataguardbroker.yaml

  dataguardbroker.database.oracle.com/dataguardbroker-sample configured
```

You can also use the patch command:

```sh
$ kubectl --type=merge -p '{"spec":{"fastStartFailover": true}}' patch dataguardbroker dataguardbroker-sample

  dataguardbroker.database.oracle.com/dataguardbroker-sample patched
```

Applying this results in the creation of a pod running the Observer. The Observer is a component of the DGMGRL interface, which monitors the availability of the primary database.

**Note:** When the attribute `fastStartFailover` is `true`, then performing a switchover by specifying `setAsPrimaryDatabase` is not allowed.

### Convert Standby to Snapshot Standby

A snapshot standby is a fully updatable standby database that can be used development and testing. It receives and archives, but does not apply redo data from a primary database. The redo data received from the primary database is applied after a snapshot standby database is converted back into a physical standby database, and after discarding all local updates to the snapshot standby database.

To convert a standby database to a snapshot standby, Ensure Fast-Start Failover is disabled, and tshen set the attribute `.spec.convertToSnapshotStandby` to `true` in [`singleinstancedatabase.yaml`](./../../config/samples/sidb/singleinstancedatabase.yaml) before applying it. For example:

```sh
$ kubectl apply -f singleinstancedatabase.yaml

  singleinstancedatabase.database.oracle.com/sidb-sample configured
```

You can also use the patch command:

```sh
$ kubectl --type=merge -p '{"spec":{"convertToSnapshotStandby":true}}' patch singleinstancedatabase sidb-sample

  singleinstancedatabase.database.oracle.com/sidb-sample patched
```

### Static Data Guard Connect String

  External and internal (running in pods) applications can always connect to the database in the primary role by using `.status.externalConnectString` and `.status.clusterConnectString` of the Data Guard broker resource respectively. These connect strings are fixed for the Data Guard broker resource, and will not change on switchover or failover. The external connect string can be obtained using the following command:

  ```sh
  $ kubectl get dataguardbroker dataguardbroker-sample -o "jsonpath={.status.externalConnectString}"

    10.0.25.87:1521/DATAGUARD
  ```
  This connect string will always automatically route to the database in the primary role. Client applications can be totally agnostic of the databases in the Oracle Data Guard configuration. Their number or host/IP details are not needed in the connect string.

### Patch Primary and Standby databases

Databases (both primary and standby) running in you cluster and managed by the Oracle Database operator can be patched between release updates of the same major release. 

To patch an existing database, edit and apply the **[`config/samples/sidb/singleinstancedatabase_patch.yaml`](../../config/samples/sidb/singleinstancedatabase_patch.yaml)** file of the database resource/object either by specifying a new release update for image attributes, or by running the following command:

```sh
kubectl --type=merge -p '{"spec":{"image":{"pullFrom":"patched-image:tag","pullSecrets":"pull-secret"}}}' patch singleinstancedatabase <database-name>

```
Follow these steps for patching databases configured with the Data Guard broker:
1. Ensure Fast-Start Failover is disabled by running the following command
```sh
  kubectl patch dataguardbroker dataguardbroker-sample -p '{"spec":{"fastStartFailover": false}}' --type=merge
```
2. Patch all the standby databases by replacing the image with the new release update image.
3. Perform switchover of the primary to one of the standby databases.
4. Patch the original primary database (currently standby after #2)  
   After step 3, the software for primary and standby databases is at the same release update
5. Bounce the current primary database by updating the replica count to 0 and then 1  
   Step 5 will trigger a datapatch execution, which results in patching the datafiles
6. Finally, perform switch over of the current primary back to the original primary (current standby)


### Delete the Data Guard Configuration

To delete a standby or primary database configured for Oracle Data Guard, delete the `dataguardbroker` resource. After that is done, delete the standby databases, and then finally the primary database.

#### Delete DataguardBroker Resource
```sh
$ kubectl delete dataguardbroker dgbroker-sample 

  dataguardbroker.database.oracle.com/dgbroker-sample deleted
```

**Note:** If a switchover to standby was performed, then ensure that you switch back to the original primary database before deleting the Data Guard broker resource. For example:
#### Delete Standby Database
```sh 
$ kubectl delete singleinstancedatabase stdby-1

  singleinstancedatabase.database.oracle.com "stdby-1" deleted
```

### Execute Custom Scripts

You can set up custom scripts (SQL, shell scripts, or both) to run after the initial database setup, and to have scripts run after each startup of the database. SQL scripts will be executed as `sysdba`, and shell scripts will be executed as the current user. To ensure proper order, Oracle recommends that you prefix your scripts with a number. For example: `01_users.sql`, `02_permissions.sql`, and son on. To ensure that these scripts are available to run after setup or after each database startup, place all such scripts in setup and startup folders created in a persistent volume.

Create a persistent volume by using [static provisioning](#static-persistence) and then specify the name of this volume with the `<.spec.persistence.scriptsVolumeName>` field which corresponds to the `scriptsVolumeName` field of the persistence section in the **[`singleinstancedatabase.yaml`](../../config/samples/sidb/singleinstancedatabase.yaml)**.

**Note:** Running custom scripts requires read and list access for persistent volumes, as mentioned in [prerequisites](#prerequisites)

## OracleRestDataService Resource

The Oracle Database Operator creates the `OracleRestDataService` as a custom resource. In this documeent, we will refer to `OracleRestDataService` as ORDS. Creating ORDS as a custom resource enables the RESTful API access to the Oracle Database in K8s, and enables it to be managed as a native Kubernetes object.

### Resource Details

#### ORDS List
To list ORDS services, use the following command: 

```sh
$ kubectl get oraclerestdataservice -o name

  oraclerestdataservice.database.oracle.com/ords-sample 

```

#### Quick Status
To obtain a quick status check of the ORDS service, use the following command:

```sh
$ kubectl get oraclerestdataservice ords-sample

NAME          STATUS      DATABASE         DATABASE API URL                                           DATABASE ACTIONS URL                           APEX URL
ords-sample   Healthy     sidb-sample      http://10.0.25.54:8181/ords/schema1/_/db-api/stable/     http://10.0.25.54:8181/ords/sql-developer     http://10.0.25.54:8181/ords/apex

```

#### Detailed Status
To obtain a detailed status check of the ORDS service, use the following command:

```sh
$ kubectl describe oraclerestdataservice ords-sample

  Name:         ords-sample
  Namespace:    default
  Labels:       <none>
  Annotations:  <none>
  API Version:  database.oracle.com/v1alpha1
  Kind:         OracleRestDataService
  Metadata: ...
  Spec: ...
  Status:
    Cluster Db API URL:    http://ords21c-1.default:8181/ords/schema1/_/db-api/stable/
    Database Actions URL:  http://10.0.25.54:8181/ords/sql-developer
    Database API URL:      http://10.0.25.54:8181/ords/schema1/_/db-api/stable/
    Apex URL:              http://10.0.25.54:8181/ords/apex
    Database Ref:          sidb21c-1
    Image:
      Pull From:     ...
      Pull Secrets:  ...
    Load Balancer:   true
    Ords Installed:  true
    Persistence:
      Access Mode:    ReadWriteMany
      Size:           100Gi
      Storage Class:  
    Service IP:       10.0.25.54
    Status:           Healthy

```

### Template YAML
  
The template `.yaml` file for Oracle Rest Data Services (`OracleRestDataService` kind), including all the configurable options, is available at **[config/samples/sidb/oraclerestdataservice.yaml](../../config/samples/sidb/oraclerestdataservice.yaml)**.

**Note:**  
- The `adminPassword` and `ordsPassword` fields in the `oraclerestdataservice.yaml` file contains secrets for authenticating the Single Instance Database and the ORDS user with the following roles: `SQL Administrator, System Administrator, SQL Developer, oracle.dbtools.autorest.any.schema`.  

- If you want to install ORDS in a [prebuilt database](#provision-a-pre-built-database), then ensure that you attach the **database persistence** by uncommenting the `persistence` section in the **[`config/samples/sidb/singleinstancedatabase_prebuiltdb.yaml`](../../config/samples/sidb/singleinstancedatabase_prebuiltdb.yaml)** file, while provisioning the prebuilt database.

### REST Enable a Database

#### Provision ORDS

To quickly provision a new ORDS instance, use the example **[`config/samples/sidb/oraclerestdataservice_create.yaml`](../../config/samples/sidb/oraclerestdataservice_create.yaml)** file. For example: 

```sh
$ kubectl apply -f oraclerestdataservice_create.yaml

  oraclerestdataservice.database.oracle.com/ords-sample created
```
After this command completes, ORDS is installed in the container database (CDB) of the Single Instance Database.

##### Note:
You are required to specify the ORDS Secret in the [`oraclerestdataservice_create.yaml`](../../config/samples/sidb/oraclerestdataservice_create.yaml) file. The default value mentioned in the `adminPassword.secretName` field is `ords-secret`. You can create this Secret manually by using the following command:

```bash
kubectl create secret generic ords-secret --from-literal=oracle_pwd=<specify password here>
```

Alternatively, you can create this Secret by filling the passwords in the **[`oraclerestdataservice_secrets.yaml`](../../config/samples/sidb/oraclerestdataservice_secrets.yaml)** file and applying it using the following command:

```bash
kubectl apply -f singleinstancedatabase_secrets.yaml
```

#### Creation Status
  
Creating a new ORDS instance takes a while. To check the status of the ORDS instance, use the following command:

```sh
$ kubectl get oraclerestdataservice/ords-sample -o "jsonpath={.status.status}"

  Healthy
```
ORDS is open for connections when the `status` column returns `Healthy`.

#### REST Endpoints

Clients can access the REST Endpoints using `.status.databaseApiUrl` as shown in the following command.

```sh
$ kubectl get oraclerestdataservice/ords-sample -o "jsonpath={.status.databaseApiUrl}"

  http://10.0.25.54:8181/ords/schema1/_/db-api/stable/
```

All the REST Endpoints can be found in [_REST APIs for Oracle Database_](https://docs.oracle.com/en/database/oracle/oracle-database/21/dbrst/rest-endpoints.html).

There are two basic approaches for authentication to the REST Endpoints. Certain APIs are specific about which authentication method they will accept.

#### Database API

To call certain REST endpoints, you must use the Schema User, which is REST-Enabled with role `SQL Administrator`, and `.spec.ordsPassword` credentials.

The Schema user also has the following additional roles: `System Administrator, SQL Developer`.

Use this Schema user to authenticate the following: 
* Database APIs
* Any Protected AutoRest Enabled Object APIs
* Database Actions of any REST Enabled Schema
  
##### Examples
Some examples for the Database API usage for REST-Enabled schema1 are as follows:
- **Get all Database Components**
    ```sh
    curl -s -k -X GET -u '<.spec.restEnableSchemas[].schemaName>:<.spec.ordsPassword>' http://10.0.25.54:8181/ords/schema1/_/db-api/stable/database/components/ | python -m json.tool
    ```
- **Get all Database Users** 
    ```sh
    curl -s -k -X GET -u '<.spec.restEnableSchemas[].schemaName>:<.spec.ordsPassword>' http://10.0.25.54:8181/ords/schema1/_/db-api/stable/database/security/users/ | python -m json.tool
    ```
- **Get all Tablespaces**
    ```sh
    curl -s -k -X GET -u '<.spec.restEnableSchemas[].schemaName>:<.spec.ordsPassword>' http://10.0.25.54:8181/ords/schema1/_/db-api/stable/database/storage/tablespaces/ | python -m json.tool
    ```
- **Get all Database Parameters**
    ```sh
    curl -s -k -X GET -u '<.spec.restEnableSchemas[].schemaName>:<.spec.ordsPassword>' http://10.0.25.54:8181/ords/schema1/_/db-api/stable/database/parameters/ | python -m json.tool
    ```
- **Get all Feature Usage Statistics**
    ```sh
    curl -s -k -X GET -u '<.spec.restEnableSchemas[].schemaName>:<.spec.ordsPassword>' http://10.0.25.54:8181/ords/schema1/_/db-api/stable/database/feature_usage/ | python -m json.tool
    ```

#### MongoDB API

To enable the Database API for MongoDB, set `.spec.mongoDbApi` to `true`. When this is done, MongoDB applications are be able to connect to Oracle Database using the MongoDB API Access URL. For example:

```sh
$ kubectl get oraclerestdataservice/ords-sample -o "jsonpath={.status.mongoDbApiAccessUrl}"

  mongodb://[{user}:{password}@]10.0.25.54:27017/{user}?authMechanism=PLAIN&authSource=$external&ssl=true&retryWrites=false&loadBalanced=true
```

* Change [{user}:{password}@] to database username and password. Retain the @ symbol but remove all the brackets.
* Change the {user} later in the URL to database username as well.

#### Advanced Usages

##### Oracle Data Pump
The Oracle REST Data Services (ORDS) database API enables you to create Oracle Data Pump export and import jobs by using REST web service calls.

REST APIs for Oracle Data Pump Jobs can be found at [https://docs.oracle.com/en/database/oracle/oracle-database/21/dbrst/op-database-datapump-jobs-post.html](https://docs.oracle.com/en/database/oracle/oracle-database/21/dbrst/op-database-datapump-jobs-post.html).
##### REST Enabled SQL

The REST-Enabled SQL functionality is available to all of the schemas specified in the `.spec.restEnableSchemas` attribute of the example yaml in the sample folder. Only these schemas will have access SQL Developer Web Console specified by the Database Actions URL. 

The REST-Enabled SQL functionality enables REST calls to send DML, DDL and scripts to any REST-Enabled schema by exposing the same SQL engine used in SQL Developer and Oracle SQLcl (SQL Developer Command Line).

For example:

**Run a Script:**

Create a file called "/tmp/table.sql" with the following contents.

```sh
  CREATE TABLE DEPT (
    DEPTNO NUMBER(2) CONSTRAINT PK_DEPT PRIMARY KEY,
    DNAME VARCHAR2(14),
    LOC VARCHAR2(13)
  ) ;

  INSERT INTO DEPT VALUES (10,'ACCOUNTING','NEW YORK');
  INSERT INTO DEPT VALUES (20,'RESEARCH','DALLAS');
  INSERT INTO DEPT VALUES (30,'SALES','CHICAGO');
  INSERT INTO DEPT VALUES (40,'OPERATIONS','BOSTON');
  COMMIT;
```

Run the following API to run the script created in the previous example:

```sh
  curl -s -k -X "POST" "http://10.0.25.54:8181/ords/<.spec.restEnableSchemas[].urlMapping>/_/sql" \
  -H "Content-Type: application/sql" \
  -u '<.spec.restEnableSchemas[].schemaName>:<.spec.ordsPassword>' \
  -d @/tmp/table.sql
```

**Basic Call:**

Fetch all entries from 'DEPT' table by calling the following API

```sh
  curl -s -k -X "POST" "http://10.0.25.54:8181/ords/<.spec.restEnableSchemas[].urlMapping>/_/sql" \
  -H "Content-Type: application/sql" \
  -u '<.spec.restEnableSchemas[].schemaName>:<.spec.ordsPassword>' \
  -d $'select * from dept;' | python -m json.tool
```

**Note:** `.spec.restEnableSchema[].urlMapping` is optional and is defaulted to `.spec.restEnableSchemas[].schemaName`

##### Database Actions

Database Actions is a web-based interface that uses Oracle REST Data Services to provide development, data tools, administration and monitoring features for Oracle Database.

* To use Database Actions, you must sign in as a database user whose schema has been REST-enabled.
* To enable a schema for REST, you can specify appropriate values for the `.spec.restEnableSchemas` attributes details in the sample `yaml` **[config/samples/sidb/oraclerestdataservice.yaml](../../config/samples/sidb/oraclerestdataservice.yaml)**, which are needed for authorizing Database Actions.
* Schema are created (if they exist) with the username as `.spec.restEnableSchema[].schema` and password as `.spec.ordsPassword.`.
* UrlMapping `.spec.restEnableSchema[].urlMapping` is optional and is defaulted to `.spec.restEnableSchema[].schema`.

Database Actions can be accessed with a browser by using `.status.databaseActionsUrl`. For example:

```sh
$ kubectl get oraclerestdataservice/ords-sample -o "jsonpath={.status.databaseActionsUrl}"

  http://10.0.25.54:8181/ords/sql-developer
```

To access Database Actions, sign in by using the following code as a database user whose schema has been REST-Enabled: 

* Login Page: \
Username: `.spec.restEnableSchemas[].schemaName` \
Password: `.spec.ordsPassword`

![database-actions-home](/images/sidb/database-actions-home.png)

For more information about Database Actions, see: [Oracle Database Actions](https://docs.oracle.com/en/database/oracle/sql-developer-web/21.2/index.html).

### APEX Installation

Oracle APEX is a low-code development platform that enables developers to build scalable, secure enterprise apps, with world-class features that can be deployed anywhere.

Using APEX, developers can quickly develop and deploy compelling apps that solve real problems and provide immediate value. Developers won't need to be an expert in a vast array of technologies to deliver sophisticated solutions. Focus on solving the problem and let APEX take care of the rest.

The `OraOperator` facilitates installation of APEX in the database and also configures ORDS for it.

* Status of APEX configuration can be checked using the following command:

  ```sh
  $ kubectl get oraclerestdataservice ords-sample -o "jsonpath={.status.apexConfigured}"

    [true]
  ```

Application Express can be accessed via browser using `.status.apexUrl` in the following command.

```sh
$ kubectl get oraclerestdataservice/ords-sample -o "jsonpath={.status.apexUrl}"

  http://10.0.25.54:8181/ords/apex
```

Sign in to Administration services using \
workspace: `INTERNAL` \
username: `ADMIN` \
password: `Welcome_1`

![application-express-admin-home](/images/sidb/application-express-admin-home.png)

**Note:**
- Oracle strongly recommends that you change the default APEX admin password.
- By default, the full development environment is initialized in APEX. After deployment, you can change it manually to the runtime environment. To change environments, run the script `apxdevrm.sql` after connecting to the primary database from the ORDS pod as the `SYS` user with `SYSDBA` privilege. For detailed instructions, see: [Converting a Full Development Environment to a Runtime Environment](https://docs.oracle.com/en/database/oracle/application-express/21.2/htmig/converting-between-runtime-and-full-development-environments.html#GUID-B0621B40-3441-44ED-9D86-29B058E26BE9).

### Delete ORDS
- To delete ORDS, run the following command:
      
      kubectl delete oraclerestdataservice ords-sample

- You cannot delete the referred Database before deleting its ORDS resource.
- APEX, if installed, also gets uninstalled from the database when ORDS gets deleted.

## Maintenance Operations
If you need to perform some maintenance operations (Database/ORDS) manually, then the procedure is as follows:
1. Use `kubectl exec` to access the pod where you want to perform the manual operation, a command similar to the following:

       kubectl exec -it <pod-name> /bin/bash

2. The important locations, such as ORACLE_HOME, ORDS_HOME, and so on, can be found in the environment, by using the `env` command.

3. Log In to `sqlplus` to perform manual operations by using the following command:       
      
        sqlplus / as sysdba

## Additional information
Detailed instructions for setting up Single Instance Database by OraOperator using OCI free trial account is available now in the LiveLab format. Please use the following link:
  [https://oracle.github.io/cloudtestdrive/AppDev/database-operator/workshops/freetier/?lab=introduction](https://oracle.github.io/cloudtestdrive/AppDev/database-operator/workshops/freetier/?lab=introduction)

Thanks, [Jan Leemans](https://github.com/janleemans), for this effort!!
