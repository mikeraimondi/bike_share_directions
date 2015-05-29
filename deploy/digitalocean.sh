curl --request POST "https://api.digitalocean.com/v2/droplets" \                                                    <<<
     --header "Content-Type: application/json" \
     --header "Authorization: Bearer $DO_TOKEN" \
     --data '{
      "region":"nyc3",
      "image":"coreos-stable",
      "size":"512mb",
      "name":"bike-share-directions-3",
      "private_networking":true,
      "ssh_keys":['$DO_KEY_ID'],
      "user_data": "'"$(cat cloud-config.yaml | sed 's/"/\\"/g')"'"
}'
