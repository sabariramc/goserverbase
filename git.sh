export message="fixed option overriding issue"
export version="v4.14.1"
export branch="v4"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch