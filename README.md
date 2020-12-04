# echo-apisvr
Echo based API Server 

[![Go Report Card](https://goreportcard.com/badge/github.com/jnpr-tjiang/echo-apisvr)](https://goreportcard.com/report/github.com/jnpr-tjiang/echo-apisvr)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CircleCI](https://circleci.com/gh/jnpr-tjiang/echo-apisvr.svg?style=shield)](https://circleci.com/gh/jnpr-tjiang/echo-apisvr)
[![codecov](https://codecov.io/gh/jnpr-tjiang/echo-apisvr/branch/main/graph/badge.svg?token=JGLAY0UYIV)](https://codecov.io/gh/jnpr-tjiang/echo-apisvr)
## Dev env setup with vscode
To run the test cases in vscode, you need to follow the steps below:
1. Follow the instruction below to setup vscode so that it can be launched from command line on Mac
https://code.visualstudio.com/docs/setup/mac#_launching-from-the-command-line
2. Launch vscode with `SCHEMA` environment variable set to the data model schema file
   ```
   $ cd /Users/tjiang/code/playground/echo-apisvr
   $ SCHEMA=/Users/tjiang/code/playground/echo-apisvr/schemas/device.json code .
   ``` 
