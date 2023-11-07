export message="Added abstraction context extraction"
export version="v3.18.3.segmentio"
export branch="segmentio"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch