export message="Streamlined request handling: removed set error in context with write error response function"
export version="v5.4.0"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch