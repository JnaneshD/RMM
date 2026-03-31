## Use the below commands to generate the Certificate

MSYS_NO_PATHCONV=1 openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
  -days 3650 -nodes \
  -subj "/CN=localhost"

## Grab your fingerprint like the below

MSYS_NO_PATHCONV=1 openssl x509 -in cert.pem -fingerprint -sha256 -noout

- Use the above fingerprint in the client code and then
- Build the client binary. Because the validation should happen both the client side and the serverside 