### Deployment sample

This examples allows you to deploy vault and secrets-manager in your own cluster, using microk8s.

1.- Deploy vault

`kubectl apply -f vault.yaml`

2.- Expose Vault port locally

`kubectl port-forward $(kubectl get po -l app=vault| awk '{print $1}' | grep -v NAME) 8200:8200`

3.- Get Vault token

`kubectl logs -l app=vault --tail=500 | grep Root`

4.- Vault setup

This will create the policy, the role and a kubernetes secret containing role_id and secret_id.

`VAULT_TOKEN=<TOKEN_FROM_STEP_3> ./vault-setup.sh`

5.- Install crd

`kubectl apply -f crd.yaml`

6.- Deploy secrets-manager

`kubectl apply -f secrets-manager.yaml`

*NOTE*: You have a `SecretDefinition` exmaple there too to play with it: `secretsmanager_v1alpha1_secretdefinition.yaml`