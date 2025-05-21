document.addEventListener('DOMContentLoaded', () => {
	initAddressSuggestions();

	// Превью логотипа
	const logoUpload = document.getElementById('logoUpload');
	const logoPreview = document.getElementById('logoPreview');

	document.querySelectorAll(".hour-select").forEach(populateHourSelect);
	document.querySelectorAll(".minute-select").forEach(populateMinuteSelect);

	// Обработчик для чекбоксов
    document.querySelectorAll('.day-active').forEach(checkbox => {
        checkbox.addEventListener('change', () => {
            const group = checkbox.closest('.working-day');
            const timeSelects = group.querySelector('.time-selects');
            if (timeSelects) {
                // Скрываем или показываем блок времени в зависимости от состояния чекбокса
                timeSelects.style.display = checkbox.checked ? 'flex' : 'none';
            }
        });
    });

	logoUpload.addEventListener('change', (e) => {
		const file = e.target.files[0];
		if (file) {
			const reader = new FileReader();
			reader.onload = (event) => {
				logoPreview.innerHTML = `<img src="${event.target.result}" style="max-width: 100%;">`;
			};
			reader.readAsDataURL(file);
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


	// Отправка формы с файлом
	if (document.getElementById('partnerForm')) {
		document.getElementById('partnerForm').addEventListener('submit', async (e) => {
			e.preventDefault();
			const formData = new FormData();

			const days = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'];
			let valid = true;

			days.forEach(day => {
				const isActive = document.querySelector(`[name=${day}_active]`).checked;
			
				if (isActive) {
					const fromHour = document.querySelector(`[name=${day}_from_hour]`)?.value;
					const fromMinute = document.querySelector(`[name=${day}_from_minute]`)?.value;
					const toHour = document.querySelector(`[name=${day}_to_hour]`)?.value;
					const toMinute = document.querySelector(`[name=${day}_to_minute]`)?.value;
			
					if (!fromHour || !fromMinute || !toHour || !toMinute) {
						alert(`Пожалуйста, заполните часы для ${day}`);
						throw new Error(`Часы не заполнены для ${day}`);
					}
			
					formData.append(`hours_${day}_from`, `${fromHour}:${fromMinute}`);
					formData.append(`hours_${day}_to`, `${toHour}:${toMinute}`);
				} else {
					formData.append(`hours_${day}_from`, `00:00`);
					formData.append(`hours_${day}_to`, `00:00`);
				}
			});

			if (!valid) return;
			console.log(document.getElementById('servicePhone').value);
			formData.append('serviceName', document.getElementById('serviceName').value);
			formData.append('serviceAddress', document.getElementById('serviceAddress').value);
			formData.append('servicePhone', document.getElementById('servicePhone').value);
			formData.append('ownerName', document.getElementById('ownerName').value);
			formData.append('ownerEmail', document.getElementById('ownerEmail').value);
			formData.append('ownerPassword', document.getElementById('ownerPassword').value);

			const logoFile = document.getElementById('logoUpload').files[0];
			if (logoFile) {
				formData.append('logo', logoFile);
			}
			const data = {
				image_url: document.getElementById('logoUpload').value
			  };
			  
			  console.log(data); 

			try {
				const response = await fetch('/api/register-partner', {
					method: 'POST',
					body: formData
				});

				if (response.ok) {
					alert('Заявка отправлена!');
					window.location.href = '/';
				} else {
					const error = await response.text();
					alert(error);
				}
			} catch (error) {
				console.error('Error:', error);
				alert('Ошибка сети');
			}
		});
	}
});

function populateHourSelect(select) {
	select.innerHTML = '<option value="" disabled selected>Часы</option>';
	for (let h = 0; h < 24; h++) {
		const hour = h.toString().padStart(2, '0');
		select.innerHTML += `<option value="${hour}">${hour}</option>`;
	}
}

function populateMinuteSelect(select) {
	select.innerHTML = '<option value="" disabled selected>Минуты</option>';
	[0, 15, 30, 45].forEach(min => {
		const minute = min.toString().padStart(2, '0');
		select.innerHTML += `<option value="${minute}">${minute}</option>`;
	});
}

// Инициализация подсказок адреса
function initAddressSuggestions() {
	const addressInput = document.getElementById('serviceAddress');
	const suggestionsContainer = document.getElementById('addressSuggestions');

	addressInput.addEventListener('input', debounce(function (e) {
		const query = e.target.value.trim();

		if (query.length < 3) {
			suggestionsContainer.classList.remove('active');
			suggestionsContainer.innerHTML = '';
			return;
		}

		// Запрос к API подсказок Яндекс.Карт
		fetch(`https://suggest-maps.yandex.ru/v1/suggest?apikey=132f0b75-64ab-4755-965f-6bad49020bdd&text=${encodeURIComponent(query)}&lang=ru_RU`)
			.then(response => {
				if (!response.ok) {
					throw new Error('Ошибка при запросе подсказок');
				}
				return response.json();
			})
			.then(data => {
				suggestionsContainer.innerHTML = '';
				suggestionsContainer.classList.add('active');
				if (data.results && data.results.length > 0) {
					data.results.forEach(item => {
						const suggestion = document.createElement('div');
						suggestion.className = 'suggestion-item';
						suggestion.textContent = item.subtitle.text + ", " + item.title.text;

						suggestion.addEventListener('click', () => {
							addressInput.value = item.subtitle.text + ", " + item.title.text;
							suggestionsContainer.classList.remove('active');
							console.log(item.title);
						});

						suggestionsContainer.appendChild(suggestion);
					});
				} else {
					suggestionsContainer.innerHTML = '<div class="suggestion-item">Нет подсказок</div>';
				}

			})
			.catch(error => {
				console.error('Ошибка:', error);
				suggestionsContainer.innerHTML = '<div class="suggestion-item">Ошибка загрузки подсказок</div>';
			});
	}, 300));

	// Скрытие подсказок при клике вне поля
	document.addEventListener('click', (e) => {
		if (!addressInput.contains(e.target) && !suggestionsContainer.contains(e.target)) {
			suggestionsContainer.classList.remove('active');
		}
	});

}


// Функция для ограничения частоты запросов
function debounce(func, wait) {
	let timeout;
	return function (...args) {
		clearTimeout(timeout);
		timeout = setTimeout(() => func.apply(this, args), wait);
	};
}