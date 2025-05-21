document.addEventListener('DOMContentLoaded', () => {
    // Элементы DOM
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const loginSection = document.getElementById('login-form');
    const registerSection = document.getElementById('register-form');
    const userInfoSection = document.getElementById('user-info');
    const userNameSpan = document.getElementById('user-name');
    const logoutBtn = document.getElementById('logout-btn');
	const adminBtn = document.getElementById('admin-btn');
	const controlBtn = document.getElementById('control-btn');
    const partnerBtn = document.getElementById('partner-btn');

    // Карусель
    const track = document.querySelector('.carousel-track');
    const slides = Array.from(document.querySelectorAll('.carousel-slide'));
    const nextBtn = document.querySelector('.next-btn');
    const prevBtn = document.querySelector('.prev-btn');
    let currentIndex = 0;
    const slideWidth = slides[0].getBoundingClientRect().width;
    let interval = setInterval(nextSlide, 5000);

    // Подключаем API Яндекс.Карт
    const script = document.createElement('script');
    script.src = getUrlKeyYMaps()
    script.onload = initMap;
    document.head.appendChild(script);

	//Переключение между формами(login/register)
	showRegister.addEventListener('click', (e) => {
		e.preventDefault();
		loginSection.style.display = 'none';
		registerSection.style.display = 'block';
	});

	showLogin.addEventListener('click', (e) => {
		e.preventDefault();
		registerSection.style.display = 'none';
		loginSection.style.display = 'block';
	});
    // Обработка входа
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('loginEmail').value;
        const password = document.getElementById('loginPassword').value;

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                credentials: 'include',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });

            if (response.ok) {
                const user = await response.json();
                updateUI(user);
            } else {
                alert('Ошибка входа');
            }
        } catch (error) {
            console.error('Ошибка входа:', error);
        }
    });

	[].forEach.call(document.querySelectorAll('.tel'), function (input) {
		var keyCode;
		function mask(event) {
			event.keyCode && (keyCode = event.keyCode);
			var pos = this.selectionStart;
			if (pos < 3) event.preventDefault();
			var matrix = "+7 (___)-___-__-__",
				i = 0,
				def = matrix.replace(/\D/g, ""),
				val = this.value.replace(/\D/g, ""),
				new_value = matrix.replace(/[_\d]/g, function (a) {
					return i < val.length ? val.charAt(i++) || def.charAt(i) : a
				});
			i = new_value.indexOf("_");
			if (i != -1) {
				i < 5 && (i = 3);
				new_value = new_value.slice(0, i)
			}
			var reg = matrix.substr(0, this.value.length).replace(/_+/g,
				function (a) {
					return "\\d{1," + a.length + "}"
				}).replace(/[+()]/g, "\\$&");
			reg = new RegExp("^" + reg + "$");
			if (!reg.test(this.value) || this.value.length < 5 || keyCode > 47 && keyCode < 58) this.value = new_value;
			if (event.type == "blur" && this.value.length < 5) this.value = ""
		}

		input.addEventListener("input", mask, false);
		input.addEventListener("focus", mask, false);
		input.addEventListener("blur", mask, false);
		input.addEventListener("keydown", mask, false)

	});

    // Обработка регистрации
    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const name = document.getElementById('registerName').value;
        const email = document.getElementById('registerEmail').value;
		const phone = document.getElementById('registerPhone').value;
        const password = document.getElementById('registerPassword').value;
        const role = document.getElementById('registerRole').value;

        try {
            const response = await fetch('/api/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, email, phone, password, role })
            });

            if (response.status === 201) {
                alert('Регистрация успешна!');
                registerSection.style.display = 'none';
                loginSection.style.display = 'block';
            } else {
                alert('Ошибка регистрации');
            }
        } catch (error) {
            console.error('Ошибка регистрации:', error);
        }
    });

    // Выход
    logoutBtn.addEventListener('click', async () => {
        try {
            await fetch('/api/logout', { method: 'GET' });
            loginSection.style.display = 'block';
            userInfoSection.style.display = 'none';
			adminBtn.style.display = 'none';
        } catch (error) {
            console.error('Ошибка выхода:', error);
        }
    });

	// Кнопка "Админ-панель"
	adminBtn.addEventListener('click', () => {
        window.location.href = '/admin';
    });

	controlBtn.addEventListener('click', () => {
		fetch('/api/partner-owner')
        .then(response => response.json())
        .then(data => {
            const serviceId = data;
            window.location.href = `/dashboard/${serviceId}`;
        })

    });



    // Кнопка "Стать партнером"
    partnerBtn.addEventListener('click', () => {
        window.location.href = '/partner-register.html';
    });

    // Проверка авторизации
    checkAuth();

    // Карусель
    function updateCarousel() {
        updateDots();
        track.style.transform = `translateX(-${currentIndex * slideWidth}px)`;
        resetInterval();
    }

    function nextSlide() {
        currentIndex = (currentIndex + 1) % slides.length;
        updateCarousel();
    }

    function prevSlide() {
        currentIndex = (currentIndex - 1 + slides.length) % slides.length;
        updateCarousel();
    }

    function resetInterval() {
        clearInterval(interval);
        interval = setInterval(nextSlide, 5000);
    }

    function updateDots() {
        document.querySelectorAll('.dot').forEach((dot, index) => {
            dot.classList.toggle('active', index === currentIndex);
        });
    }

    nextBtn.addEventListener('click', nextSlide);
    prevBtn.addEventListener('click', prevSlide);
    track.addEventListener('mouseenter', () => clearInterval(interval));
    track.addEventListener('mouseleave', resetInterval);


    function renderServices(services) {
        const container = document.getElementById('services-list');
        container.innerHTML = services.map(service => `
            <div class="service-card">
                <h3>${service.name}</h3>
                <p>${service.address}</p>
                <p>Телефон: ${service.phone}</p>
            </div>
        `).join('');
    }

    // Функции авторизации
    async function checkAuth() {
        try {
            const response = await fetch('/api/user');
            if (response.ok) {
                const user = await response.json();
                updateUI(user);
            }
        } catch (error) {
            console.error('Ошибка проверки авторизации:', error);
        }
    }

    function updateUI(user) {
        loginSection.style.display = 'none';
        registerSection.style.display = 'none';
        userInfoSection.style.display = 'block';
        userNameSpan.textContent = user.Name;

		// Показываем кнопку "Админ-панель" только для admin
        if (user.Role === 'admin') {
            adminBtn.style.display = 'block';
        } else {
            adminBtn.style.display = 'none';
        }
		if (user.Role === 'admin_service') {
			controlBtn.style.display = 'block';
		} else {
            controlBtn.style.display = 'none';
        }
    }

    // Инициализация карты
    function initMap() {
        ymaps.ready(() => {
            const map = new ymaps.Map('map', {
                center: [55.796, 49.106], // Москва
                zoom: 12
            });

            const clusterer = new ymaps.Clusterer({
                preset: 'islands#invertedBlueClusterIcons'
            });

            fetch('/api/partners')
                .then(res => {
                    if (!res.ok) throw new Error('Ошибка загрузки партнеров');
                    return res.json();
                })
                .then(partners => {
                    const partnersList = document.getElementById('partners-list');
                    let placemarks = [];

                    partners.forEach(partner => {
                        const item = document.createElement('div');
                        item.className = 'partner-card';
                        item.dataset.id = partner.id;
                        item.innerHTML = `
                            <div class="partner-image-container${!partner.logoPath ? ' placeholder' : ''}">
                                ${partner.logoPath ? 
                                    `<img src="/${partner.logoPath}" alt="${partner.name}" class="partner-image">` : 
                                    'Логотип'}
                            </div>
                            <div class="partner-info">
                                <h3 class="partner-name">${partner.name}</h3>
                                <div class="partner-meta">
                                    <p class="partner-address">📍 ${partner.address}</p>
                                    <p class="partner-phone">📞 ${partner.phone}</p>
                                </div>
                            </div>
                        `;

                        if (partner.latitude && partner.longitude) {
                            const placemark = new ymaps.Placemark(
                                [partner.latitude, partner.longitude],
                                {
                                    hintContent: partner.name,
                                    balloonContent: `
                                        <strong>${partner.name}</strong><br>
                                        ${partner.address}<br>
                                        📞 ${partner.phone}
                                    `,
                                    partnerId: partner.id
                                },
                                {
                                    preset: 'islands#blueDotIcon'
                                }
                            );
                            placemarks.push(placemark);

                            // Наведение на карточку
                            item.addEventListener('mouseover', () => {
                                document.querySelectorAll('.partner-card').forEach(el => el.classList.remove('active'));
                                item.classList.add('active');
                                if (partner.latitude && partner.longitude) {
                                    map.setCenter([partner.latitude, partner.longitude], 15);
                                    placemark.options.set('preset', 'islands#redDotIcon');
                                }
                            });

                            // Уход курсора с карточки
                            item.addEventListener('mouseout', () => {
                                item.classList.remove('active');
                                placemark.options.set('preset', 'islands#blueDotIcon');
                            });

                            // Клик по карточке
                            item.addEventListener('click', () => {
                                window.location.href = `/partner/${partner.id}`;
                            });
                        }

                        partnersList.appendChild(item);
                    });

                    clusterer.add(placemarks);
                    map.geoObjects.add(clusterer);
                })
                .catch(error => {
                    console.error('Ошибка загрузки партнеров:', error);
                });
        });
    }
});