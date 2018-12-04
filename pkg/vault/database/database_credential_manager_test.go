package database

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appscode/pat"
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubedb/apimachinery/apis/authorization/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/secret/engines/database"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
)

const (
	credResponse = `
{
   "lease_id":"1204",
   "lease_duration":300,
   "data":{
      "username":"nahid",
      "password":"1234"
   }
}
`
)

func vaultServer() *httptest.Server {
	m := pat.New()

	m.Get("/v1/database/creds/geterror", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("error"))
	}))
	m.Get("/v1/database/creds/jsonerror", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("json error"))
	}))
	m.Get("/v1/database/creds/success", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(credResponse))
	}))
	m.Put("/v1/sys/leases/revoke/success", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	m.Put("/v1/sys/leases/revoke/error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("error"))
	}))
	m.Put("/v1/sys/leases/lookup", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			LeaseID string `json:"lease_id"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}

		if data.LeaseID == "1234" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"errors":["invalid lease"]}`))
		}

	}))

	return httptest.NewServer(m)
}

func TestCreateSecret(t *testing.T) {
	cred := &vaultapi.Secret{
		LeaseID:       "1204",
		LeaseDuration: 300,
		Data: map[string]interface{}{
			"password": "1234",
			"username": "nahid",
		},
	}

	dbAreq := &api.DatabaseAccessRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "db-req",
			UID:  "1234",
		},
	}

	testData := []struct {
		testName     string
		dClient      *DBCredManager
		cred         *vaultapi.Secret
		secretName   string
		namespace    string
		createSecret bool
	}{
		{
			testName: "Successfully secret created",
			dClient: &DBCredManager{
				dbAccessReq: dbAreq,
				vaultClient: nil,
			},
			cred:         cred,
			secretName:   "pg-cred",
			namespace:    "pg",
			createSecret: false,
		},
		{
			testName: "create secret, secret already exists, no error",
			dClient: &DBCredManager{
				dbAccessReq: dbAreq,
				vaultClient: nil,
			},
			cred:         cred,
			secretName:   "pg-cred",
			namespace:    "pg",
			createSecret: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			d := test.dClient
			d.kubeClient = kfake.NewSimpleClientset()

			if test.createSecret {
				_, err := d.kubeClient.CoreV1().Secrets(test.namespace).Create(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: test.namespace,
						Name:      test.secretName,
					},
					Data: map[string][]byte{
						"test": []byte("hi"),
					},
				})

				assert.Nil(t, err)
			}

			err := d.CreateSecret(test.secretName, test.namespace, test.cred)
			if assert.Nil(t, err) {
				_, err := d.kubeClient.CoreV1().Secrets(test.namespace).Get(test.secretName, metav1.GetOptions{})
				assert.Nil(t, err)
			}
		})
	}
}

func TestCreateRole(t *testing.T) {
	dbAreq := &api.DatabaseAccessRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "db-req",
			UID:  "1234",
		},
	}

	testData := []struct {
		testName   string
		dClient    *DBCredManager
		createRole bool
		roleName   string
		secretName string
		namespace  string
	}{
		{
			testName: "Successfully role created",
			dClient: &DBCredManager{
				dbAccessReq: dbAreq,
				vaultClient: nil,
			},
			createRole: false,
			roleName:   "pg-role",
			secretName: "pg-cred",
			namespace:  "pg",
		},
		{
			testName: "create role, role already exists, no error",
			dClient: &DBCredManager{
				dbAccessReq: dbAreq,
				vaultClient: nil,
			},
			createRole: true,
			roleName:   "pg-role",
			secretName: "pg-cred",
			namespace:  "pg",
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			d := test.dClient
			d.kubeClient = kfake.NewSimpleClientset()

			if test.createRole {
				_, err := d.kubeClient.RbacV1().Roles(test.namespace).Create(&rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      test.roleName,
						Namespace: test.namespace,
					},
				})

				assert.Nil(t, err)
			}

			err := d.CreateRole(test.roleName, test.namespace, test.secretName)
			if assert.Nil(t, err) {
				r, err := d.kubeClient.RbacV1().Roles(test.namespace).Get(test.roleName, metav1.GetOptions{})
				if assert.Nil(t, err) {
					assert.Equal(t, "", r.Rules[0].APIGroups[0], "api group")
					assert.Equal(t, "secrets", r.Rules[0].Resources[0], "resources")
					assert.Equal(t, test.secretName, r.Rules[0].ResourceNames[0], "resource name")
					assert.Equal(t, "get", r.Rules[0].Verbs[0], "verbs")
				}
			}
		})
	}
}

func TestCreateRoleBinding(t *testing.T) {
	dbAreq := &api.DatabaseAccessRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "db-req",
			UID:  "1234",
		},
	}

	subs := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Namespace: "pg",
			Name:      "pg-sa",
		},
	}

	testData := []struct {
		testName        string
		dClient         *DBCredManager
		createRB        bool
		roleName        string
		roleBindingName string
		namespace       string
		subjects        []rbacv1.Subject
	}{
		{
			testName: "Successfully role binding created",
			dClient: &DBCredManager{
				dbAccessReq: dbAreq,
				vaultClient: nil,
			},
			createRB:        false,
			roleName:        "pg-role",
			roleBindingName: "pg-role-binding",
			subjects:        subs,
		},
		{
			testName: "Successfully role binding patched",
			dClient: &DBCredManager{
				dbAccessReq: dbAreq,
				vaultClient: nil,
			},
			createRB:        true,
			roleName:        "pg-role",
			roleBindingName: "pg-role-binding",
			subjects:        subs,
			namespace:       "pg",
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			d := test.dClient
			d.kubeClient = kfake.NewSimpleClientset()

			if test.createRB {
				_, err := d.kubeClient.RbacV1().RoleBindings(test.namespace).Create(&rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      test.roleBindingName,
						Namespace: test.namespace,
					},
				})

				assert.Nil(t, err)
			}

			err := d.CreateRoleBinding(test.roleBindingName, test.namespace, test.roleName, test.subjects)
			if assert.Nil(t, err) {
				r, err := d.kubeClient.RbacV1().RoleBindings(test.namespace).Get(test.roleBindingName, metav1.GetOptions{})
				if assert.Nil(t, err) {
					assert.Equal(t, test.roleName, r.RoleRef.Name, "role ref role name")
					assert.Equal(t, "Role", r.RoleRef.Kind, "role ref role kind")
					assert.Equal(t, rbacv1.GroupName, r.RoleRef.APIGroup, "role ref role api group")
				}
			}
		})
	}
}

func TestRevokeLease(t *testing.T) {
	srv := vaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL

	cl, err := vaultapi.NewClient(cfg)
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}

	dbAreq := &api.DatabaseAccessRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "db-req",
			UID:  "1234",
		},
	}

	testData := []struct {
		testName    string
		dClient     *DBCredManager
		expectedErr bool
		leaseID     string
	}{
		{
			testName: "Lease revoke successful",
			dClient: &DBCredManager{
				dbAccessReq: dbAreq,
				vaultClient: cl,
			},
			leaseID:     "success",
			expectedErr: false,
		},
		{
			testName: "Lease revoke failed",
			dClient: &DBCredManager{
				dbAccessReq: dbAreq,
				vaultClient: cl,
			},
			leaseID:     "error",
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			err := test.dClient.RevokeLease(test.leaseID)
			if test.expectedErr {
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDatabaseRoleBinding_IsLeaseExpired(t *testing.T) {
	srv := vaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL

	cl, err := vaultapi.NewClient(cfg)
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}

	testData := []struct {
		testName  string
		dClient   *DBCredManager
		isExpired bool
		leaseID   string
	}{
		{
			testName: "lease is expired",
			dClient: &DBCredManager{
				vaultClient: cl,
			},
			isExpired: true,
			leaseID:   "1222",
		},
		{
			testName: "lease is valid",
			dClient: &DBCredManager{
				vaultClient: cl,
			},
			isExpired: false,
			leaseID:   "1234",
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			ok, err := test.dClient.IsLeaseExpired(test.leaseID)
			if assert.Nil(t, err) {
				assert.Condition(t, func() (success bool) {
					if ok == test.isExpired {
						return true
					}
					return false
				})
			}
		})
	}
}

func TestDBCredManager_GetCredential(t *testing.T) {
	srv := vaultServer()
	defer srv.Close()

	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL

	cl, err := vaultapi.NewClient(cfg)
	if !assert.Nil(t, err, "failed to create vault client") {
		return
	}

	cred := &vaultapi.Secret{
		LeaseID:       "1204",
		LeaseDuration: 300,
		Data: map[string]interface{}{
			"password": "1234",
			"username": "nahid",
		},
	}

	dbAreq := &api.DatabaseAccessRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "db-req",
			UID:  "1234",
		},
	}

	testData := []struct {
		testName    string
		dClient     *DBCredManager
		cred        *vaultapi.Secret
		expectedErr bool
	}{
		{
			testName: "Successfully get credential",
			dClient: &DBCredManager{
				dbAccessReq:  dbAreq,
				vaultClient:  nil,
				secretGetter: database.NewSecretGetter(cl, "database", "success"),
			},
			cred:        cred,
			expectedErr: false,
		},
		{
			testName: "failed to get credential, json error",
			dClient: &DBCredManager{
				dbAccessReq:  dbAreq,
				vaultClient:  nil,
				secretGetter: database.NewSecretGetter(cl, "database", "jsonerror"),
			},
			cred:        nil,
			expectedErr: true,
		},
		{
			testName: "failed to get credential, get error",
			dClient: &DBCredManager{
				dbAccessReq:  dbAreq,
				vaultClient:  nil,
				secretGetter: database.NewSecretGetter(cl, "database", "geterror"),
			},
			cred:        nil,
			expectedErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			d := test.dClient
			cred, err := d.GetCredential()
			if test.expectedErr {
				assert.NotNil(t, err, "error should occur")
			} else {
				if assert.Nil(t, err) {
					assert.Equal(t, *test.cred, *cred, "credential should match")
				}
			}
		})
	}
}
