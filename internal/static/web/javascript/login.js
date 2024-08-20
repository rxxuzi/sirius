document.getElementById('login-form').addEventListener('submit', function(e) {
    e.preventDefault();
    const formData = new FormData(this);
    const jsonData = {};
    formData.forEach((value, key) => {jsonData[key] = value});

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