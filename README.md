# NCCPA Checker

## Install

`go install github.com/pma9/nccpa-checker`

## Configuration

You need to first setup a Twilio account and add funds there.
Twilio is the service that will send you text messages.

Create a `.env` file in the same directory where nccpa-checker is

```
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
TWILIO_FROM_NUMBER=
TWILIO_TO_NUMBER=
```

Now you need to get a token that will be used for your requests.
This token can be found on the NCCPA website. Send a request and
check the network tab of the developer tools. The request body
of the "portal.nccpa.net" request will have a token. Copy this token
and save it as `token.txt` in the same directory as the nccpa-checker.

## Usage

By ID

`nccpa-checker --id <nccpa id>`

By Params

`nccpa-checker --fn <First Name> --ln <Last Name>`

Defaults:
- State : California (CA)
- Country : United States (USA)

`nccpa-checker --fn <First Name> --ln <Last Name> --sc <State Code> --cc <Country Code>`

You can also specify a different token file

`nccpa-checker ... --tf <token file path>`