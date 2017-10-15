docker ps -qa | awk '{print $1}' | while read AA; do test -n "$AA" && docker kill $AA; done
docker ps -qa | awk '{print $1}' | while read AA; do test -n "$AA" && docker rm $AA; done
docker images | grep "^dev" | awk '{print $1}' | while read AA; do test -n "$AA" && docker rmi -f $AA; done
docker network prune -f
