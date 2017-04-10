# Vault-Chef

Issues Hashicorp Vault tokens to Chef clients after authentication

The golang version of [smart_proxy_vault](https://github.com/theforeman/smart_proxy_vault).

## Parameters

This application is configured through environment variables. All standard Vault environment variables will be respected.

#### CHEF_SERVER

The URL of your Chef server.

#### CHEF_VERIFY_TLS

Toggles whether or not to verify the certificate of your Chef server. Defaults to true.

#### VAULT_ADDR

The URL to your Vault server. Defaults to "127.0.0.1:8200".

#### VAULT_TOKEN

The VAULT_TOKEN for the application. This can be wrapped or unwrapped, and it should be a role token. The token must have the appropriate policies assigned to it:

    * One for creating role tokens
    * One for reading: SECRET_CERT_PATH, and SECRET_CHEF_PEM_PATH
    * Optional: One for lookup-self, and renew-self (if RENEW_TOKEN is set to true)

#### VAULT_SKIP_VERIFY

Toggles whether or not to verify the certificate of your Vault server. Defaults to false.

#### SECRET_CERT_PATH

The path in Vault at which a public and private key resides. They will be used to serve this application over HTTPS. The key names in Vault must be "cert" and "key".

#### SECRET_CHEF_PEM_PATH

The path in Vault at which a privileged Chef client pem resides to be used for communicating with the Chef server's API. The key name in Vault must be "key".

#### RENEW_TOKEN

Toggles whether or not to setup a go routine to periodically renew the Vault token. The appropriate policies must be on the token for this. Defaults to true.

#### LISTEN_ADDR

The IP and port to serve this application over. Defaults to ":8443"


### Example

```
VAULT_TOKEN="TOKEN" CHEF_SERVER="https://chef.example.com" CHEF_CLIENT_NAME="my-chef-client-name" CHEF_VERIFY_TLS=true VAULT_ADDR="https://vault.service.consul" VAULT_SKIP_VERIFY=false SECRET_CERT_PATH="secret/vault-chef/chef.example.com" SECRET_CHEF_PEM_PATH="secret/vault-chef/chef" RENEW_TOKEN=true LISTEN_ADDR="127.0.0.1:8443" ./vault-chef
2017/04/10 08:19:36 [INFO] Starting vault-chef server, listening on 127.0.0.1:8443
```

## Endpoints

#### POST `/token/issue/:role`

`:role` should be the name of a token role on the Vault server

*body*

Your request body can specify any of the parameters the direct [Vault endpoint](https://www.vaultproject.io/docs/auth/token.html#auth-token-create-role_name-) accepts.


## Chef Authentication

In order for a Chef client to successfully authenticate, it must define certain headers in its POST request.

#### X-VAULT-CLIENT

This header should contain the Chef client ID (e.g. `knife client [name]`). This is used by the [go-chef/chef](https://github.com/go-chef/chef) package to fetch a client.

#### X-VAULT-SIGNATURE

This header should be a Base64 encoded signature with the body containing the client's public key. Here's a Ruby example of how to generate this on your client nodes:

```ruby
require 'openssl'
require 'base64'

def sign_request(key_path)
  rsa = OpenSSL::PKey::RSA.new File.read key_path
  Base64.strict_encode64(rsa.sign(OpenSSL::Digest::SHA512.new, rsa.public_key.to_s))
end

signature = sign_request('/etc/chef/client.pem')
```