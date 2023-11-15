docker build --pull --rm -f "Dockerfile_occult" -t cirrusli/x_im:occult-v1.0 "."
docker build --pull --rm -f "Dockerfile_gateway" -t cirrusli/x_im:gateway-v1.0 "."
docker build --pull --rm -f "Dockerfile_logic" -t cirrusli/x_im:logic-v1.0 "."
docker build --pull --rm -f "Dockerfile_router" -t cirrusli/x_im:router-v1.0 "."

docker push cirrusli/x_im:occult-v1.0
docker push cirrusli/x_im:gateway-v1.0
docker push cirrusli/x_im:logic-v1.0
docker push cirrusli/x_im:router-v1.0
