# Simple Secrets Manager Example

This is an example guide to perform a simple deployment of secrets-manager in K8s. It will deploy _secrets-manager_ in `default` namespace. It also includes a dummy Vault deployment but you can use your own if you already have a Vault instance. 


## Deploy Vault

If you already have a Vault instance, you can skip this step.

```bash
kubectl apply -f vault.yaml
```

## Generate a Vault token 
... and store it in a K8s secret. If you are going to use your current Vault instance you don't need the first line, just generate a Vault token as usual and go straight to second line to create it.

```bash
VAULT_TOKEN=$(kubectl exec $(kubectl get pods -l app=vault -o custom-columns=:metadata.name) -c vault -- vault token create --field=token)

kubectl create secret generic --from-literal=token=$VAULT_TOKEN vault-token-secret

```

## Deploy Secrets Manager

The K8s manifests in [secrets-manager.yaml](secrets-manager.yaml) will configure to _secrets-manager_ to fetch secrets from the dummy Vault deployment, but you can change the configmap content to use your own secrets.

```
kubectl apply -f secrets-manager.yaml
```

If you check the _secrets-manager_ logs, you will see after a while that it's updating the K8s secrets with the content from Vault.

```
➜  kubectl logs $(kubectl get pods -l app=demo-secrets-manager -o custom-columns=:metadata.name)
time="2018-11-28T00:15:00Z" level=info msg="successfully logged in to Vault cluster vault-cluster-0587b7b9"
time="2018-11-28T00:15:15Z" level=info msg="secret 'default/supersecret1' must be updated"
time="2018-11-28T00:15:15Z" level=info msg="secret 'default/supersecret1' updated"
time="2018-11-28T00:15:15Z" level=info msg="secret 'default/supersecret2' must be updated"
time="2018-11-28T00:15:15Z" level=info msg="secret 'default/supersecret2' updated"
```

```
➜  kubectl get secrets
NAME                  TYPE                                  DATA      AGE
default-token-9k5dv   kubernetes.io/service-account-token   3         34d
supersecret1          kubernetes.io/tls                     2         22h
supersecret2          Opaque                                2         22h
vault-token-secret    Opaque                                1         22h
```

