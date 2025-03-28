#!/bin/bash

if [[ $SQL_DSN == "" ]]; then
  echo "Missing SQL_DSN variable"
  exit 1
fi

psql "$SQL_DSN" -f ./scripts/sql/tables.sql