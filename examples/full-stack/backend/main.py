# Simple mock backend server on port 8000
from http.server import HTTPServer, BaseHTTPRequestHandler
import json

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        response = {"message": "Hello from Backend API on port 8000 via PORTA!", "path": self.path}
        self.wfile.write(json.dumps(response).encode())

if __name__ == '__main__':
    print("Backend listening on http://localhost:8000")
    server = HTTPServer(('127.0.0.1', 8000), Handler)
    server.serve_forever()
