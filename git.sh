export message="Streamlined config: removed probagation of service"
export version="v5.1.3"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch``