#!/bin/bash

vaultenv() {
    local path
    local file
    path=$1
    file=$(venv "$path")
    dotenv "$file"
}
