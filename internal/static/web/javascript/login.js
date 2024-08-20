document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const formData = new FormData(this);
            const jsonData = {
                host: formData.get('host'),
                port: parseInt(formData.get('port'), 10),
                user: formData.get('username'),
                pass: formData.get('password')
            };

            if (isNaN(jsonData.port)) {
                document.getElementById('message').textContent = 'Error: Port must be a valid number';
                return;
            }

            fetch('/connect', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(jsonData)
            })
                .then(response => response.text())
                .then(data => {
                    if (data === "Connected successfully") {
                        window.location.href = '/';
                    } else {
                        document.getElementById('message').textContent = data;
                    }
                })
                .catch(error => {
                    document.getElementById('message').textContent = 'Error: ' + error;
                });
        });
    } else {
        console.error('Login form not found');
    }
});