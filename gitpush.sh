#!/bin/bash

DATE=`date +'%D %T'`

git rm -r --cached .
git add .
git commit -m "$DATE"
git push origin main --force