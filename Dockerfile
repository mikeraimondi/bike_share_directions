FROM scratch

COPY app/backend/bike_share_directions /
COPY dist /dist

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/bike_share_directions"]
