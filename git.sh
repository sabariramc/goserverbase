export message="Fixed error code not proccesed for custom error"
export version="v4.7.3"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch