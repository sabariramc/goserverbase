export message="BaseApp: added GetNotifier and update log print flow for PanicRecovery and ProcessError method"
export version="v5.0.22"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch``