# Use a base image with both Go and Python
FROM golang:1.20 as builder

# Set the working directory
WORKDIR /app

# Copy the Go source code into the container
COPY app/binary.go .

# Build the Go binary
RUN go build -o app binary.go

# Use a Python base image for the final stage
FROM python:3.11

RUN apt-get update && apt-get install -y
RUN pip install --no-cache-dir --target=/app ebooklib lxml bs4 cmudict ollama

    # calibre \
    # pandoc \
    # && rm -rf /var/lib/apt/lists/*

# RUN apt-get install flite -y

# RUN tar zxvf flite-2.3-current.tar.gz
# RUN cd flite-2.3-current
# RUN git clone http://github.com/festvox/flite

# COPY flite /flite

# RUN cd /flite/
# RUN ./configure && make
# RUN sudo make install
# RUN cd testsuite
# RUN make lex_lookup
# RUN sudo cp lex_lookup /usr/local/bin

# COPY flite/testsuite/lex_lookup /app/lex_lookup



# Set the working directory
WORKDIR /app

# Copy the built Go binary from the builder stage
COPY --from=builder /app/app /app/app

# Copy the Python script into the container
COPY app/script.py /app/script.py

COPY heteronyms.json /app/heteronyms.json

# Copy any additional required files, like the input file
# COPY input input
COPY app/input /app/input
# COPY output output
COPY app/output /app/output

# Expose the port used by the Go binary
EXPOSE 8080

# Install any Python dependencies (if needed)
# RUN pip install -r requirements.txt

# Start the Go binary and the Python script
CMD ["/bin/sh", "-c", "./app & python script.py"]
