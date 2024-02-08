export message="updated dependency version"
export version="v3.19.0"
export branch="v3"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $branch