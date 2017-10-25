**calChecker** checks your Google Calendar for any appointments for today.

### OAuth
To use calChecker you must set up a Google API OAuth Application.  This is very easy to do if you already have a google account.

[OAuth Tutorial](/OAuth.md)

### Usage
```bash
$ calChecker --credentialFile {downloaded_file} --tokenFile token.json
```
The `credentialFile` is downloaded at the end of the [OAuth Tutorial](/OAuth.md).

The `tokenFile` will be created for you.

### First Time Setup
You should see a message like this:
```bash
Attempting to open https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=********.apps.googleusercontent.com&redirect_uri=http%3A%2F%2F127.0.0.1%3A37345&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fcalendar.readonly&state=state-token in your browser
```

If your browser doesn't automatically open you can copy and paste the link.

Note: **You must open the link on the same computer.**

![Authorize Access](https://raw.githubusercontent.com/guywithnose/calChecker/master/images/authorize.png)

Once the app is authorized it will show your appointments for today.
```bash
$ calChecker --credentialFile {downloaded_file} --tokenFile token.json
Wed, 11:30AM  Dentist
Wed, 3:00PM   Meet Bob
```
