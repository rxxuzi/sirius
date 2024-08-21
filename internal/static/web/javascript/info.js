function fetchStaticInfo() {
    fetch('/api/static-info')
        .then(response => response.text())
        .then(data => {
            document.getElementById('static-info').textContent = data;
        })
        .catch(error => {
            console.error('Error fetching static info:', error);
            document.getElementById('static-info').textContent = 'Error fetching system information.';
        });
}

function fetchRealTimeInfo() {
    fetch('/api/live-info')
        .then(response => response.json())
        .then(data => {
            document.getElementById('cpu-usage').textContent = data.CPU;
            document.getElementById('memory-usage').textContent = data.Memory;
            document.getElementById('gpu-usage').textContent = data.GPU;
        })
        .catch(error => {
            console.error('Error fetching real-time info:', error);
        });
}

document.getElementById('refresh-btn').addEventListener('click', () => {
    fetchStaticInfo();
    fetchRealTimeInfo();
});
fetchStaticInfo();
fetchRealTimeInfo();
setInterval(fetchRealTimeInfo, 1000);