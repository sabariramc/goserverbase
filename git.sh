export message="Added abstraction for kafka writer"
export version="v3.18.2.segmentio"
export branch="segmentio"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch