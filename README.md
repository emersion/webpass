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

You can also choose not to store your encrypted PGP private key on the server,
in this case you'll have to carry it with you e.g. on a USB stick.

## Configuration

Create `config.json`:

```json
{
	"auth": {
		"type": "git",
		"url": "git@git.example.org:user/pass-store.git",
		"privatekey": "/home/user/.ssh/id_rsa"
	},
	"pgp": {
		"privatekey": "private-key.gpg"
	}
}
```

* `auth`: configures authentication. `auth.type` must be one of:
	* `none`: no authentication. You should configure HTTP authentication with
	a reverse proxy for instance.
	* `pam`: uses the current user's account.
	* `git`: uses a remote Git repository, which is cloned in memory when logging
	in. The repository's URL must be specified with `auth.url`, and a SSH
	private key can be specified with `auth.privatekey`.
* `pgp`: configures OpenPGP
	* `pgp.privatekey`: path to your OpenPGP private key. If not specified, your
	private key will be requested when decrypting a password.

## Security

Once logged in, the encrypted PGP key and the encrypted passwords will be served
by the API.

The PGP key password won't be sent to the server, since the passwords are
decrypted client-side.

## License

MIT
