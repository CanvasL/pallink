FROM swaggerapi/swagger-ui:v5.32.0

COPY deploy/swagger/swagger.json /usr/share/nginx/html/swagger/swagger.json
RUN chmod 644 /usr/share/nginx/html/swagger/swagger.json

ENV BASE_URL=/docs
ENV SWAGGER_JSON=/usr/share/nginx/html/swagger/swagger.json
ENV PERSIST_AUTHORIZATION=true
