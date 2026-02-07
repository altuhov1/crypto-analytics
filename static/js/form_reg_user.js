function togglePassword() {
    const passwordInput = document.getElementById('password');
    const toggleBtn = passwordInput.nextElementSibling;

    if (passwordInput.type === 'password') {
        passwordInput.type = 'text';
        toggleBtn.textContent = '🔒';
    } else {
        passwordInput.type = 'password';
        toggleBtn.textContent = '👁️';
    }
}

// Валидация формы
document.getElementById('loginForm').addEventListener('submit', function (e) {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    // Проверка на пустые поля
    if (!username || !password) {
        e.preventDefault();
        alert('Пожалуйста, заполните все обязательные поля!');
        return;
    }

    // Проверка на опасные SQL символы - УНИВЕРСАЛЬНАЯ
    const dangerousPatterns = /['"\\]|--|\/\*|\*\/|union|select|insert|update|delete|drop|create|alter|exec|script|<|>/gi;
    // Универсальная проверка любого поля
    function checkField(fieldValue, fieldName) {
        const lines = fieldValue.split('\n');
        for (let line of lines) {
            if (dangerousPatterns.test(line)) {
                return true;
            }
        }
        return false;
    }

    if (checkField(username, 'username') || checkField(password, 'password')) {
        e.preventDefault();
        const errorElement = document.getElementById('errorMessage');
        errorElement.innerHTML = `
            <span class="error-icon">⚠️</span>
            Обнаружены недопустимые символы в данных
        `;
        errorElement.classList.add('has-icon');
        errorElement.style.display = 'flex';
        return;
    }

    // Дополнительная проверка длины
    if (username.length > 50 || password.length > 100) {
        e.preventDefault();
        const errorElement = document.getElementById('errorMessage');
        errorElement.innerHTML = `
            <span class="error-icon">⚠️</span>
            Слишком длинные данные
        `;
        errorElement.classList.add('has-icon');
        errorElement.style.display = 'flex';
        return;
    }
});

// Обработка query параметров
document.addEventListener('DOMContentLoaded', function () {
    const urlParams = new URLSearchParams(window.location.search);
    const errorType = urlParams.get('err');

    const errorElement = document.getElementById('errorMessage');

    if (errorType === 'password') {
        errorElement.innerHTML = `
                    <span class="error-icon">⚠️</span>
                    Неправильно введен пароль. Пожалуйста, попробуйте снова.
                `;
        errorElement.classList.add('has-icon');
        errorElement.style.display = 'flex';
        document.getElementById('password').focus();

    } else if (errorType === 'nilUser') {
        errorElement.innerHTML = `
                    <span class="error-icon">👤</span>
                    Пользователя с таким логином не существует
                `;
        errorElement.classList.add('has-icon');
        errorElement.style.display = 'flex';
        document.getElementById('username').focus();
    }
});