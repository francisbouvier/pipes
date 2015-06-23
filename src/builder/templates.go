package builder

const T_GENERIC = `
FROM <BASE_IMAGE>

ADD <EXEC_PATH_SRC> <EXEC_PATH_DEST>

# Get the pipes_client
ADD pipes_client bin/pipes_client
RUN chmod 755 bin/pipes_client

ENTRYPOINT ["pipes_client"]
`
