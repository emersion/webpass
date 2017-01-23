# webpass

A web interface for [pass](https://www.passwordstore.org/), a UNIX password manager.

## Usage

```shell
go get -u github.com/emersion/webpass/...

cd $GOPATH/src/github.com/emersion/webpass
npm install
gpg --export-secret-keys > private-key.gpg

webpass
```

Go to http://localhost:8080. You'll be first asked for your login password.
Once logged in, a list of your passwords is displayed. When you click on an
item, your PGP key password will be prompted and the password will be displayed.

You can now setup an HTTPS reverse proxy to webpass.

## Security

Once logged in, the encrypted PGP key and the encrypted passwords will be served
by the API.

The PGP key password won't be sent to the server, since the passwords are
decrypted client-side.

## License

MIT
