const pwShowHide = document.getElementById('pw_hide');

document.querySelectorAll('.pw_hide').forEach(icon => {
    icon.addEventListener('click', () => {
        let getPwInput = icon.parentElement.querySelector('input');
        if (getPwInput.type === 'password') {
            getPwInput.type = 'text';
            icon.classList.replace('uil-eye-slash', 'uil-eye');
        } else {
            getPwInput.type = 'password';
            icon.classList.replace('uil-eye', 'uil-eye-slash');
        }
    });
});