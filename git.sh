export message="Kafka producer and consumer config changes"
export version="v3.5.0"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push github $version
git push origin $branch
git push github $branch