#!/bin/bash

vaultenv() {
    local path
    local file
    path=$1
    file=$(vaultenv "$path")
    dotenv "$file"
}
