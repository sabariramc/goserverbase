export message="Updated dependency version"
export version="v4.14.3"
export branch="v4"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch