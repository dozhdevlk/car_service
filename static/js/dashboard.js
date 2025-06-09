var partnerId;
let allBookings = [];

document.addEventListener('DOMContentLoaded', () => {
	//  Извлечение partnerId из пути URL (например, /dashboard/1)
	const pathSegments = window.location.pathname.split('/').filter(segment => segment);
	partnerId = parseInt(pathSegments[pathSegments.length - 1]);
	if (!partnerId || isNaN(partnerId)) {
		console.error('Не удалось извлечь ID сервиса из URL');
		return;
	}


	// Отображаем ID сервиса в заголовке
	const partnerIdElement = document.getElementById('partner-id');
	if (partnerIdElement) {
		partnerIdElement.textContent = `#${partnerId}`;
	}

	document.getElementById('search-button').addEventListener('click', () => {
		const searchId = document.getElementById('search-id').value;
		const searchPhone = document.getElementById('search-phone').value;
		const selectedStatus = document.getElementById('status-filter').value;
		applyFilters(selectedStatus, searchId, searchPhone);
	});

	// Обработчик добавления/редактирования объявления
	const addAnnouncementBtn = document.getElementById('addAnnouncementBtn');
	addAnnouncementBtn.addEventListener('click', () => {
		document.getElementById('announcementId').value = '';
		document.getElementById('announcementTitle').value = '';
		document.getElementById('announcementText').value = '';
		document.getElementById('announcementImage').value = '';
		new bootstrap.Modal(document.getElementById('editAnnouncementModal')).show();
	});

	// Отправка формы добавления/редактирования объявления
	const editAnnouncementForm = document.getElementById('editAnnouncementForm');
	editAnnouncementForm.addEventListener('submit', function (e) {
		e.preventDefault();

		const id = document.getElementById('announcementId').value;
		const title = document.getElementById('announcementTitle').value;
		const text = document.getElementById('announcementText').value;
		const image = document.getElementById('announcementImage').files[0];

		const formData = new FormData();
		formData.append('title', title);
		formData.append('text', text);
		formData.append('partner_id', parseInt(partnerId));
		if (image) {
			formData.append('image_url', image);
		}

		const data = {
			partner_id: partnerId,
			title: document.getElementById('announcementTitle').value,
			text: document.getElementById('announcementText').value,
			image_url: document.getElementById('announcementImage').value
		};

		console.log(data);  // Проверьте, что выводится в консоль

		let url = '/api/announcements';
		let method = 'POST';
		if (id) {
			url += `/${partnerId}/${id}`;
			method = 'PUT';
		}

		fetch(url, {
			method: method,
			body: formData,
		})
			.then(response => response.json())
			.then(data => {
				if (data.error) {
					alert(data.error);
				} else {
					alert('Объявление успешно добавлено/отредактировано!');
					loadAnnouncements(partnerId); // Обновление списка объявлений
					bootstrap.Modal.getInstance(document.getElementById('editAnnouncementModal')).hide();
				}
			})
			.catch(error => {
				console.error('Ошибка добавления/редактирования объявления:', error);
				alert('Ошибка при добавлении/редактировании объявления.');
			});
	});

	// Загрузка объявлений
	function loadAnnouncements(partnerId) {
		fetch(`/api/announcements/${partnerId}`)
			.then(response => response.json())
			.then(data => {
				const announcementsList = document.getElementById('announcements-list');
				announcementsList.innerHTML = ''; // Очистить перед обновлением
				if (!data) {
					announcementsList.innerHTML = '<tr><td colspan="4">Объявлений нет</td></tr>';
					return;
				}

				data.forEach(announcement => {
					const row = document.createElement('tr');
					row.innerHTML = `
                    <td><img src="/${announcement.image_url}" alt="${announcement.title}" width="50"></td>
                    <td>${announcement.title}</td>
                    <td>${announcement.text}</td>
                    <td>
                        <button class="btn btn-warning btn-sm edit-btn" data-id="${announcement.id}">Редактировать</button>
                        <button class="btn btn-danger btn-sm delete-btn" data-id="${announcement.id}">Удалить</button>
                    </td>
                `;
					announcementsList.appendChild(row);
				});

				// Обработчики для кнопок редактирования и удаления
				document.querySelectorAll('.edit-btn').forEach(button => {
					button.addEventListener('click', function () {
						const id = this.getAttribute('data-id');
						console.log('ID объявления:', this.getAttribute('data-id'));
						fetch(`/api/announcements/${partnerId}/${id}`)
							.then(response => response.json())
							.then(data => {
								console.log('Данные объявления:', data);
								document.getElementById('announcementId').value = data.id;
								document.getElementById('announcementTitle').value = data.title;
								document.getElementById('announcementText').value = data.text;
								new bootstrap.Modal(document.getElementById('editAnnouncementModal')).show();
							});
					});
				});

				document.querySelectorAll('.delete-btn').forEach(button => {
					button.addEventListener('click', function () {
						const id = this.getAttribute('data-id');
						if (confirm('Вы уверены, что хотите удалить это объявление?')) {
							fetch(`/api/announcements/${partnerId}/${id}`, {
								method: 'DELETE',
							})
						}
					});
				});
			})
			.catch(error => {
				console.error('Ошибка загрузки объявлений:', error);
			});
	}

	loadAnnouncements(partnerId); // Загружаем объявления при загрузке страницы
	loadServices(partnerId)
	loadServiceInfo(partnerId)


	document.querySelectorAll('.tab-link').forEach(link => {
		link.addEventListener('click', (e) => {
			e.preventDefault();
			document.querySelectorAll('.tab-link').forEach(l => l.classList.remove('active'));
			document.querySelectorAll('.main-tab-content').forEach(c => {
				c.style.display = 'none';
			});

			link.classList.add('active');
			const tabId = link.getAttribute('data-tab');
			const tabContent = document.getElementById(tabId);
			tabContent.style.display = 'block';

			if (tabId === 'orders-content') {
				loadAllBookings(partnerId);
			}
		});
	});

	document.querySelector('.tab-link').click();


	// Обработка добавления услуги
	const addServiceButton = document.getElementById('add-service');
	if (addServiceButton) {
		addServiceButton.addEventListener('click', () => {
			const serviceName = document.getElementById('service-name').value;
			const servicePrice = document.getElementById('service-price').value;
			const serviceImage = document.getElementById('service-image').files[0];

			if (!serviceName || !servicePrice) {
				showServiceMessage('Пожалуйста, заполните название и цену.', 'error');
				return;
			}

			const formData = new FormData();
			formData.append('name', serviceName);
			formData.append('price', servicePrice);
			formData.append('partner_id', partnerId);
			if (serviceImage) {
				formData.append('image', serviceImage);
			}

			fetch('/api/partner_offerings', {
				method: 'POST',
				body: formData,
			})
				.then(response => response.json())
				.then(data => {
					if (data.error) {
						showServiceMessage(data.error, 'error');
					} else {
						showServiceMessage('Услуга успешно добавлена!', 'success');
						document.getElementById('service-name').value = '';
						document.getElementById('service-price').value = '';
						document.getElementById('service-image').value = '';
						loadServices(partnerId); // Перезагружаем список услуг
					}
				})
				.catch(error => {
					console.error('Ошибка добавления услуги:', error);
					showServiceMessage('Не удалось добавить услугу. Попробуйте снова.', 'error');
				});
		});
	}
	// Обработчик открытия модального окна для редактирования информации
	const editServiceInfoBtn = document.getElementById('edit-service-info-btn');
	if (editServiceInfoBtn) {
		editServiceInfoBtn.addEventListener('click', () => {
			fetch(`/api/partner/${partnerId}`)
				.then(response => {
					if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
					return response.json();
				})
				.then(data => {
					document.getElementById('editServiceInfoId').value = partnerId;
					document.getElementById('editServiceInfoName').value = data.name || '';
					document.getElementById('editServiceInfoPhone').value = data.phone || '';
					document.getElementById('editServiceInfoDescription').value = data.description || '';
					new bootstrap.Modal(document.getElementById('editServiceInfoModal')).show();
				})
				.catch(error => console.error('Ошибка загрузки данных:', error));
		});
	}

	// Обработчик сохранения изменений в модальном окне
	const editServiceInfoForm = document.getElementById('editServiceInfoForm');
	if (editServiceInfoForm) {
		editServiceInfoForm.addEventListener('submit', (e) => {
			e.preventDefault();
			const newName = document.getElementById('editServiceInfoName').value.trim();
			const newPhone = document.getElementById('editServiceInfoPhone').value.trim();
			const newDescription = document.getElementById('editServiceInfoDescription').value.trim();

			if (!newName || !newPhone) {
				showServiceMessage('Название и телефон обязательны.', 'error');
				return;
			}

			updateServiceInfo(partnerId, newName, newPhone, newDescription);
			bootstrap.Modal.getInstance(document.getElementById('editServiceInfoModal')).hide();
		});
	}
	function loadServiceInfo(partnerId) {
		fetch(`/api/partner/${partnerId}`)
			.then(response => {
				if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
				return response.json();
			})
			.then(data => {
				document.getElementById('current-service-name').textContent = data.name || 'Не указано';
				document.getElementById('current-service-phone').textContent = data.phone || 'Не указан';
				document.getElementById('current-service-description').textContent = data.description || 'Не указано';
			})
			.catch(error => {
				console.error('Ошибка загрузки информации:', error);
				showServiceMessage('Не удалось загрузить информацию.', 'error');
			});
	}
	// Функция для обновления информации о сервисе
	function updateServiceInfo(partnerId, name, phone, description) {
		fetch(`/api/partner/${partnerId}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name, phone, description }),
		})
			.then(response => {
				if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
				return response.json();
			})
			.then(data => {
				showServiceMessage('Информация успешно обновлена!', 'success');
				loadServiceInfo(partnerId);
			})
			.catch(error => {
				console.error('Ошибка обновления информации:', error);
				showServiceMessage('Не удалось обновить информацию.', 'error');
			});
	}

	function showServiceMessage(message, type) {
		const messageElement = document.getElementById('service-message');
		messageElement.textContent = message;
		messageElement.className = type === 'error' ? 'error-message' : 'success-message';
		messageElement.style.display = 'block';
		setTimeout(() => {
			messageElement.style.display = 'none';
		}, 5000);
	}

});

function loadAllBookings(partnerId) {
	fetch(`/api/bookings/${partnerId}`)
		.then(response => response.json())
		.then(bookings => {
			allBookings = bookings;
			applyFilters('all');
		})
		.catch(error => {
			console.error('Ошибка загрузки записей:', error);
			document.getElementById('all-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
		});
}

function applyFilters(status = 'all', searchId = '', searchPhone = '') {
	const filteredBookings = allBookings.filter(booking => {
		const isStatusMatch = (status === 'all' || booking.status === status);
		const isIdMatch = searchId ? booking.id.toString().includes(searchId) : true;
		const isPhoneMatch = searchPhone ? booking.user_phone.includes(searchPhone) : true;
		return isStatusMatch && isIdMatch && isPhoneMatch;
	});

	updateTabContent(filteredBookings);
}

function updateTabContent(bookings) {
	const list = document.getElementById('all-list');
	list.innerHTML = '';

	if (bookings.length === 0) {
		list.innerHTML = '<p>Записей не найдено.</p>';
	} else {
		bookings.forEach(booking => {
			const bookingCard = document.createElement('div');
			bookingCard.className = 'booking-card';
			bookingCard.innerHTML = `
                <div class="booking-info">
                    <p><strong>ID записи:</strong> ${booking.id}</p>
                    <p><strong>Название партнера:</strong> ${booking.partner_name}</p>
                    <p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
                    <p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
                    <p><strong>Имя пользователя:</strong> ${booking.user_name}</p>
                    <p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
                    <p><strong>Email пользователя:</strong> ${booking.user_email}</p>
                    <p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
                    <p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
                    <p><strong>Статус:</strong> ${booking.status}</p>
                </div>
                <div class="booking-actions">
                    ${getActionButton(booking.status, booking.id)}
                </div>
            `;
			list.appendChild(bookingCard);
		});
	}
}

function getActionButton(status, bookingId) {
	switch (status) {
		case '⏳ Ожидает подтверждения':
			return `
                <button onclick="updateBookingStatus(${bookingId}, '✅ Подтверждена')">Подтвердить</button>
                <button onclick="updateBookingStatus(${bookingId}, '❌ Отменена')">Отменить</button>
            `;
		case '✅ Подтверждена':
			return `
                <button onclick="updateBookingStatus(${bookingId}, '🔧 В работе')">В работу</button>
            `;
		case '🔧 В работе':
			return `
                <button onclick="updateBookingStatus(${bookingId}, '🏁 Завершена')">Завершить</button>
            `;
		default:
			return '';
	}
}


function updateBookingStatus(bookingId, status) {
	fetch(`/api/bookings/${bookingId}`, {
		method: 'PUT',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify({ status }),
	})
		.then(response => response.json())
		.then(data => {
			if (data.error) {
				alert(data.error);
			} else {
				alert(`Запись успешно обновлена`);
				loadAllBookings(partnerId);
			}
		})
		.catch(error => {
			console.error('Ошибка обновления статуса:', error);
			alert('Не удалось обновить статус записи.');
		});
}

// Функция загрузки услуг
function loadServices(partnerId) {
	fetch(`/api/partner_offerings?partner_id=${partnerId}`)
		.then(response => {
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			return response.json();
		})
		.then(data => {
			const servicesBody = document.getElementById('servicesBody');
			servicesBody.innerHTML = ''; // Очистка таблицы

			if (data.length === 0) {
				servicesBody.innerHTML = '<tr><td colspan="4">Услуг нет</td></tr>';
				return;
			}

			data.forEach(service => {
				const row = document.createElement('tr');
				row.innerHTML = `
                    <td>${service.id}</td>
                    <td>${service.name}</td>
                    <td>${service.price}</td>
                    <td>
                        <button class="btn btn-warning btn-sm edit-btn" data-id="${service.id}" data-name="${service.name}" data-price="${service.price}">Изменить</button>
                        <button class="btn btn-danger btn-sm delete-btn" data-id="${service.id}">Удалить</button>
                    </td>
                `;
				servicesBody.appendChild(row);
			});

			// Обработчики событий для кнопок "Изменить"
			document.querySelectorAll('.edit-btn').forEach(button => {
				button.addEventListener('click', function () {
					const id = this.getAttribute('data-id');
					const name = this.getAttribute('data-name');
					const price = this.getAttribute('data-price');

					document.getElementById('editServiceId').value = id;
					document.getElementById('editServiceName').value = name;
					document.getElementById('editServicePrice').value = price;

					const modal = new bootstrap.Modal(document.getElementById('editServiceModal'));
					modal.show();
				});
			});

			// Обработчики событий для кнопок "Удалить"
			document.querySelectorAll('.delete-btn').forEach(button => {
				button.addEventListener('click', function () {
					const id = this.getAttribute('data-id');
					if (confirm('Вы уверены, что хотите удалить эту услугу?')) {
						fetch(`/api/partner_offerings/${id}`, {
							method: 'DELETE',
							headers: {
								'Content-Type': 'application/json',
							}
						})
							.then(response => {
								if (!response.ok) {
									throw new Error(`HTTP error! status: ${response.status}`);
								}
								return response.json();
							})
							.then(data => {
								alert(data.message); // Успешное удаление
								loadServices(partnerId); // Обновляем таблицу
							})
							.catch(error => console.error('Ошибка удаления услуги:', error));
					}
				});
			});
		})
		.catch(error => console.error('Ошибка загрузки услуг:', error));
}

// Функция для отправки запроса PUT
document.getElementById('editServiceForm').addEventListener('submit', function (e) {
	e.preventDefault();

	const id = document.getElementById('editServiceId').value;
	const name = document.getElementById('editServiceName').value;
	const price = document.getElementById('editServicePrice').value;

	if (id) {
		// Обновление существующей услуги
		fetch(`/api/partner_offerings/${id}`, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({
				name: name,
				price: parseFloat(price)
			})
		})
			.then(response => {
				if (!response.ok) {
					throw new Error(`HTTP error! status: ${response.status}`);
				}
				return response.json();
			})
			.then(data => {
				alert(data.message); // Успешное обновление
				const modal = bootstrap.Modal.getInstance(document.getElementById('editServiceModal'));
				modal.hide();
				loadServices(partnerId); // Обновляем таблицу
			})
			.catch(error => console.error('Ошибка обновления услуги:', error));
	} else {
		// Создание новой услуги
		fetch('/api/partner_offerings', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({
				name: name,
				price: parseFloat(price),
				partner_id: partnerId
			})
		})
			.then(response => {
				if (!response.ok) {
					throw new Error(`HTTP error! status: ${response.status}`);
				}
				return response.json();
			})
			.then(data => {
				alert(data.message || 'Услуга успешно добавлена'); // Успешное добавление
				const modal = bootstrap.Modal.getInstance(document.getElementById('editServiceModal'));
				modal.hide();
				loadServices(partnerId); // Обновляем таблицу
			})
			.catch(error => console.error('Ошибка добавления услуги:', error));
	}
});

// Добавление новой услуги
document.getElementById('addServiceBtn').addEventListener('click', function () {
	document.getElementById('editServiceId').value = '';
	document.getElementById('editServiceName').value = '';
	document.getElementById('editServicePrice').value = '';

	const modal = new bootstrap.Modal(document.getElementById('editServiceModal'));
	modal.show();
});

function showServiceMessage(message, type) {
	const messageElement = document.getElementById('service-message');
	messageElement.textContent = message;
	messageElement.className = type === 'error' ? 'error-message' : 'success-message';
	messageElement.style.display = 'block';
	setTimeout(() => {
		messageElement.style.display = 'none';
	}, 5000);
}