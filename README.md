# lti-1-3-go-library
Proof of concept golang library for lti 1.3

Based on IMS Global public documentation and [PHP reference implementation](https://github.com/IMSGlobal/lti-1-3-php-library)

## Prepare

```bash
# assumes the project is at: ${GOPATH}/src/github.com/GRT/lti-1-3-go-library
$ cd ${GOPATH}/src/github.com/GRT/lti-1-3-go-library
# fetch all the dependencies
$ go get ./...
```

## GRT Go Test Platform

Run the tool within the test/reference platform: https://lti-ri.imsglobal.org/platforms/163/

### Running the Tool

```bash
$ cd ${GOPATH}/src/github.com/GRT/lti-1-3-go-library
$ go run *.go
2019/09/17 09:57:48 FsStore created, maxLen: 65536, root dir: /var/folders/v4/pc39k0950y7d919wxcc9xllc0000gn/T/
2019/09/17 09:57:48 Starting...

```

To perform an LTI 1.3 Tool launch:
* In a browser, go to reference app Platform home (https://lti-ri.imsglobal.org/platforms/163/)
* Click Resource Links
* Click the 'Select User for Launch' button (under 'GRT Go Test Tool Instructor' resource link)
* Choose an existing user and click 'Launch Resource Link (OIDC)'
* Click 'Post Request' (You should see a log entry from the running go app as it handles the OIDC request)
* Click the 'Launch Resource Link' button
* Click the 'Perform Launch' button (on the bottom of the page that shows what the platform will send to the tool).
* The tool launches, returning a page that echoes back much of the launch data from the platform.

From the above point, to exercise the Name and Role Provisioning Services:
* Click on the 'Fetch Members' button
* This should load a couple of new rows to the bottom of the table, including the actual JSON response as well as a table of members

Continuing the example to use the Assignment and Grade Services:
* Choose a user in the 'Member Grading' row
* Enter a grade (0-100) for one of the users and hit the 'Send Grade' Button
* You should see a popup containing the JSON response as well as a new row with the label 'Last Grade Response' that shows the same JSON

Lastly to fetch the grades from the Assignment and Grade Service:
* Push the 'Fetch Grades' Button
* This should load a new row at the bottom of the table that shows both the JSON response of all the grades, and a tabular display of the grades

Note: Deep Linking is not yet implemented in the POC.

### Keys
Published for demonstration purposes only.  
These keys are registered with the platform.

#### Public
```text
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAw20enLu/5pKsVufby15V
+GPF2iGXVhGI+HjddRi964YrpZ9QOIE9y4jTTryGiSey61aAiQpNEOsagZDmEJej
o3DGm7YKt7cAjpfU1iyA2UwdpTY/dj6v3tk6hN5qMf6mS4TZc5vkKlHOO8YJIYvl
0CSgrIkyh91IbAF4dz5ge3H1j/lPQfCfHNEPcaUsa4PKaq+kQKJ7zdeijjidBDtp
HiuJ0WL9i/dCctnSGQMBlGfF+AEqKlR7dMxzZW8G/lFYloDVdExkQhTffknYsBVo
BTBDVhiStAJj4BOXX0yRFBpCHKwngLrFX4MlhME54UD33TuJ/6bekd3SMaeDLDTe
RQIDAQAB
-----END PUBLIC KEY-----
```
#### Private
```text
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAw20enLu/5pKsVufby15V+GPF2iGXVhGI+HjddRi964YrpZ9Q
OIE9y4jTTryGiSey61aAiQpNEOsagZDmEJejo3DGm7YKt7cAjpfU1iyA2UwdpTY/
dj6v3tk6hN5qMf6mS4TZc5vkKlHOO8YJIYvl0CSgrIkyh91IbAF4dz5ge3H1j/lP
QfCfHNEPcaUsa4PKaq+kQKJ7zdeijjidBDtpHiuJ0WL9i/dCctnSGQMBlGfF+AEq
KlR7dMxzZW8G/lFYloDVdExkQhTffknYsBVoBTBDVhiStAJj4BOXX0yRFBpCHKwn
gLrFX4MlhME54UD33TuJ/6bekd3SMaeDLDTeRQIDAQABAoIBAQDBiSRC5fDDIMiJ
/G6Adfk/11dOkeu08gKqx8/RsmILiMFa8W5ZtyyCkAtsM2otFGKti+oZTVlLAvoq
DFy7W+FT8FDQTjVJAXQMwzKltEcaa7YEMrggmy5CzPCWO0oCHwWDTpqnUmBgxMkw
Cwxp64j5W/y/QdQIF04soDw0I1Mbi+HvLWySHHrJBzMNmcYyz2G4QCK8twyEAsZB
Xptlspj69PSrYvcJW7R/wjwP0edMXrKekg3I+acFiVHcxanWszMT7AorKo6vRWEZ
Xnk6PDlAc/PzelpBKC+8ShYNOgAKMjHfcn6uKAmc+o6DAR+LTKhYKdbGl+TowaDh
ugwvZP+VAoGBAOe320O3GFfl5cXbGSyOHBfmVvZYtHFO7F6ow596jnJmaIaSZQsk
6f/AAV5CIzsFQ0U/nWcsOm0SzaP9FwpHYw7hQJwdAa6ejHoirMHEv5jMzbPW6d3X
6Dccp37I0SOJclNZcITQji0buhCwOB0TFbEF22YCzyv6hTianknRC26rAoGBANfn
ry2oqvhJP6ALhmtEY8NCHAr8O3Svbey0uJ7/m5MKvkLe+ON4z230VqkWfBj4pWFm
rgxSZAlQRaUoyU5NThLGTEIBguCsHItOs5b1OM4Duiasj5u7dgX7DNKrSNAfNwiI
T5LcYzOEqFFTj5E3l8BrLGkjZVtdAtHYyN5QmybPAoGBAIUIn5Ae/JDqYqLXiXp1
FFf8XI0OnHo5L6ehCL704/d2KCiqv+xIAzhcCe0N16A5A0gsn7fuQpUAqKOv2JyE
I7EVTbzTQnX4fPpaEgklZkLZwnevuZEuNhn+D4PQ05Gthb+op9r4ycfIFWkjYvP+
UwPRMwc8Mak0KWw4CQykQgYFAoGBAMVe6BaeVUVKeN6PCp+u0mBiZA5qzNN7t8qm
3wuC8a63KH0rJm2UOFP1BO/oPSP60fy7iCp9ezPEbRZxta1eIBwrqPTCOum2jRWQ
qb47iGVUpOzL3TBpa5hGC0/fA1422vFy4wOHcyxafiByehEvuAtQLjYjBHpECdra
Ca6qE1ujAoGADsd3GqFyxvMlhALNhItjo82r0U5WJ4xMg6FunVN51+vE1zLo8Jnp
oNXNkez29COENrwoKuguglfeWkgvBl9hzPeYntzcBhAWsFJLap+sBv4iQ66++xR7
aw8ABRe3yiw0W2k8ieNJNy1un5jZz3gOQ7oDNfFrzN2S349ub+eMdzA=
-----END RSA PRIVATE KEY-----
```

## This Test Tool

Published for demonstration purposes only.  

The Public key below is registered with the Platform.  The private key is in our [registration datastore](registrationDatastore/registrations.json)

### Keys
#### Public
```text
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0PL/U28fBwe8gHhLzkth
0nPl+9VMJ7g+5jflrrjci1zPgSPANGllFH+9yew/z+CulxRQ116uCdREvfS61Isf
HhSlHLr0yjL97hez6zv+sxlSDK/hWoh2dMC6owLqO1xQtqqINj+rXqpqSABetrs5
ft0AvYCqzIKk06WCzGMk+F5RMH1Yzc6zS/Lsbvp2aIZu5OKe0WWo/5I3NgZUj3OF
YfPB2xmA1+kXPQIdOTEtkzb1oc0/xgkv0OQ3edhQsgmQPrbZmfUsKGY0+MnjMrBy
6DbFginB9staSqmDyvF1j6rYc81M7EfGFtPcGlP2PuceFXQohME4wLQIymyAOliE
zQIDAQAB
-----END PUBLIC KEY-----
```
#### Private
```text
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0PL/U28fBwe8gHhLzkth0nPl+9VMJ7g+5jflrrjci1zPgSPA
NGllFH+9yew/z+CulxRQ116uCdREvfS61IsfHhSlHLr0yjL97hez6zv+sxlSDK/h
Woh2dMC6owLqO1xQtqqINj+rXqpqSABetrs5ft0AvYCqzIKk06WCzGMk+F5RMH1Y
zc6zS/Lsbvp2aIZu5OKe0WWo/5I3NgZUj3OFYfPB2xmA1+kXPQIdOTEtkzb1oc0/
xgkv0OQ3edhQsgmQPrbZmfUsKGY0+MnjMrBy6DbFginB9staSqmDyvF1j6rYc81M
7EfGFtPcGlP2PuceFXQohME4wLQIymyAOliEzQIDAQABAoIBACxWA1tezsSdHaBc
5ijl0eHn+brP7ZLYA3CyF6hVTWa80MLkJRp56prI6Cp6WKfxUtp30xd/3Yn0Yomz
7hi/VGD7nHVWLi7hVwQ4P1MArfCuxLwwba7aGdh4NKH2MmFaGz5HPRPVurUhj9+r
RG2dmHuUxV1wec1fQz4tdm2L9AJHDrLRTpF3iAWVyIhz0XOKh+LIuMefeXSWs0FF
6uDRrkCckC4UfoxfRXNAF/3UWOOb+RA/zK31RXB+YgLUi0imVVwQNctC08r3+lOy
yxHi2aykuq/lUxPGxMR60fmmd65RNLgmzwHljEZJ7M73wWDYZ6Az57Bjp294+xXT
uOkV0YECgYEA9uxOkvk2wCLputwVMlb7P90iKFpntIeby69sbVf0pCy8ud28tfcE
GY1RTmaJFNW573oHDW9gVTmGZu3CjRO1aItW7OYJXtoFvHrheDZ6rs+At0c4e/qj
zTbes4ZAvsAVhUtTkyGNI3m5NX6GnGnSMvjVb+3TljCFt6CxrsP+lb0CgYEA2KFV
cZM9F0JCoOi2nlKjrbxfdCOOaVRuAoT/E/mFAjqTUirbnwL0frBveAGFQMNKM7Io
2n5jRxmiUyhL2T7wTYGRJ11ZWtc8ZsBO3rsBdG4o575GO55hWKlSANLNVaHNKN2F
fINA+7fB0vVoiPxoc4N/+r9LKFdmH7NhOZIP9FECgYBfCE2pZT78LbO1FhUWXcGv
L6WA0GKPaY29k9NwNeTS9uDfzAZgJiSuzOPY/7+MhEFeeKGUOyRhSJWAscspzscH
6HDZFiPPHKwOgWCbiqQm+Xe5kjCcDrfSOGb3wxjSEU13Eqmku8n9OFDe1MZsFpIu
yfQjcu33JM+h/7fC4m3uJQKBgGZ8ZSz3SJahXV482nCqjg8aqFoMnEpOjEEa5IZx
rLByP9JGvmJLBpqNJB81MPKDsa4lYliEJLm1ces/jCq6MPuqCZ8C9cwZOdUus+GB
vV105FtG1HlOI6XLbSVAla4mfyYPLyDKA8tSkxsXyR3NtCi6FKjvKUJrnr/uoFeZ
N30RAoGBANTz+kAaq0wdV5osIG00tnBRSEAErdTXEITvHhtfYjFzG2cc89LcbM/6
0icUS0EZPVLNEJC0YZnvw1zwqKPSnhBDNHK8tS3R9OraI/pBS2Bsp0bOmDTC7k0I
y2dW9ahYzAFM3G1e4nw2Capkog/TbAACpezYVijjpgH4kb7pY/Eu
-----END RSA PRIVATE KEY-----
```

## Moodle as the Platform 

TBD.