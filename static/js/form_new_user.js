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

function toggleConfirmPassword() {
    const confirmInput = document.getElementById('confirmPassword');
    const toggleBtn = confirmInput.nextElementSibling;

    if (confirmInput.type === 'password') {
        confirmInput.type = 'text';
        toggleBtn.textContent = '🔒';
    } else {
        confirmInput.type = 'password';
        toggleBtn.textContent = '👁️';
    }
}

// Валидация паролей
document.getElementById('registerForm').addEventListener('submit', function (e) {
    const username = document.getElementById('username').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirmPassword').value;

    // Проверка на пустые поля
    if (!username || !email || !password || !confirmPassword) {
        e.preventDefault();
        alert('Пожалуйста, заполните все обязательные поля!');
        return;
    }

    // Проверка на опасные SQL символы - УНИВЕРСАЛЬНАЯ
    const dangerousPatterns = /['"\\]|--|\/\*|\*\/|union|select|insert|update|delete|drop|create|alter|exec|script|<|>/gi;
    function checkField(fieldValue) {
        const lines = fieldValue.split('\n');
        for (let line of lines) {
            if (dangerousPatterns.test(line)) {
                return true;
            }
        }
        return false;
    }

    // ОДНА проверка вместо трех!
    if (checkField(username) || checkField(email) || checkField(password)) {
        e.preventDefault();
        showError('⚠️ Обнаружены недопустимые символы в данных');
        return;
    }

    // Дополнительная проверка длины
    if (username.length > 50 || email.length > 100 || password.length > 100) {
        e.preventDefault();
        showError('⚠️ Слишком длинные данные');
        return;
    }

    // Проверка email формата
    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailPattern.test(email)) {
        e.preventDefault();
        showError('⚠️ Введите корректный email адрес');
        return;
    }

    // Проверка паролей
    if (password !== confirmPassword) {
        e.preventDefault();
        showError('⚠️ Пароли не совпадают');
        return;
    }

    // Проверка сложности пароля
    if (password.length < 8 || password.length > 72) {
        e.preventDefault();
        showError('⚠️ Пароль должен содержать от 8 до 72 символов');
        return;
    }
});

// Проверка query параметров при загрузке страницы
document.addEventListener('DOMContentLoaded', function () {
    const urlParams = new URLSearchParams(window.location.search);
    const errorType = urlParams.get('err');

    if (errorType === 'alreadyExistsName') {
        showError('❌ Это имя пользователя уже занято. Пожалуйста, выберите другое.');
    } else if (errorType === 'alreadyExistsAcc') {
        showError('❌ Аккаунт с этой почтой уже существует. Попробуйте войти или восстановить пароль.');
    }
});