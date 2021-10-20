#!/bin/sh
export VAULT_ADDR=http://localhost:8200
echo "Waiting vault to launch on 8200..."

while ! nc -z localhost 8200; do   
  sleep 0.1 # wait for 1/10 of the second before check again
done

echo "Vault launched"

echo "Enabling approle"
vault auth enable approle

echo "Creating vault policy"
cat > secrets-manager.hcl  <<EOF
path "secret/data/*" {
  capabilities = ["read"]
}
EOF

cat secrets-manager.hcl | vault policy write secrets-manager -

echo "creating role"

vault write auth/approle/role/secrets-manager policies=secrets-manager secret_id_num_uses=0 secret_id_ttl=0

echo "creating some secrets"

vault kv put secret/pathtosecret1 "value=dmFsdWUzCg=="
vault kv put secret/pathtosecret2 "value=value2"
vault kv put secret/pathtosecret3 "value=value3"


echo "creating approle secret"
kubectl delete secret vault-approle-secret 2>/dev/null || true
kubectl create secret generic vault-approle-secret --from-literal role_id=$(vault read --field role_id auth/approle/role/secrets-manager/role-id) --from-literal secret_id=$(vault write --field secret_id -force auth/approle/role/secrets-manager/secret-id)
