FROM python

ADD entities.py bin/entities.py

# Get the pipes_client
ADD https://s3-us-west-1.amazonaws.com/pipesdocker/pipes_client bin/pipes_client
RUN chmod 755 bin/pipes_client

ENTRYPOINT ["bin/entities.py"]