export message="Kafka auto commit changes"
export version="v3.16.1"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch