document.getElementById('contactForm').addEventListener('submit', function (e) {
    const name = document.getElementById('name').value;
    const email = document.getElementById('email').value;
    const message = document.getElementById('message').value;

    // Проверка на пустые поля
    if (!name || !email || !message) {
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

    // Проверяем все поля
    if (checkField(name) || checkField(email) || checkField(message)) {
        e.preventDefault();
        alert('⚠️ Обнаружены недопустимые символы в данных');
        return;
    }

    // Дополнительная проверка длины
    if (name.length > 100 || email.length > 100 || message.length > 2000) {
        e.preventDefault();
        alert('⚠️ Слишком длинные данные');
        return;
    }

    // Проверка email формата
    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailPattern.test(email)) {
        e.preventDefault();
        alert('⚠️ Введите корректный email адрес');
        return;
    }
});