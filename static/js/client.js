document.getElementById("connectTelegramBtn").addEventListener("click", async () => {
	const response = await fetch("/api/telegram/init", { method: "POST" });
	if (response.ok) {
		const data = await response.json();
		const botLink = `https://t.me/InfoCarService_bot?start=${data.token}`;
		window.open(botLink, "_blank"); // Открываем Telegram-бота
	} else {
		alert("Ошибка при генерации токена.");
	}
});

const pathSegments = window.location.pathname.split('/').filter(segment => segment);
const tabs = document.querySelectorAll('.tab');
tabs.forEach(tab => {
	tab.addEventListener('click', () => {
		tabs.forEach(t => t.classList.remove('active'));
		document.querySelectorAll('.order-tab-content').forEach(content => {
			content.classList.remove('active');
			content.style.display = 'none';
		});

		tab.classList.add('active');
		const tabId = tab.getAttribute('data-tab');
		const tabContent = document.getElementById(tabId);
		tabContent.classList.add('active');
		tabContent.style.display = 'block';

		// Перезагружаем записи при переключении вкладок заказов
		if (['pending', 'confirmed', 'canceled', 'working', 'end'].includes(tabId)) {
			setTimeout(() => {
				loadBookings();
			}, 0);
		}
	});
});

loadBookings()
loadUserInfo()

function loadUserInfo() {
	fetch(`/api/user`)
		.then(response => {
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			return response.json();
		})
		.then(userInfo => {
			document.getElementById('client-name').textContent = userInfo.Name;
			document.getElementById('client-phone').textContent = userInfo.Phone;
			document.getElementById('client-email').textContent = userInfo.Email;
			if (userInfo.Tg) {
				document.getElementById('connectTelegramBtn').style.display = 'none'
			}
		})
}
function loadBookings() {
	fetch(`/api/bookings-client`)
		.then(response => {
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			return response.json();
		})
		.then(bookings => {
			const canceledList = document.getElementById('canceled-list');
			const workingList = document.getElementById('working-list');
			const endList = document.getElementById('end-list');

			pendingList.innerHTML = '';
			confirmedList.innerHTML = '';
			canceledList.innerHTML = '';
			workingList.innerHTML = '';
			endList.innerHTML = '';

			const canceledBookings = bookings.filter(booking => booking.status === '❌ Отменена');
			const workingBookings = bookings.filter(booking => booking.status === '🔧 В работе' || booking.status === '⏳ Ожидает подтверждения'
			);
			const endBookings = bookings.filter(booking => booking.status === '🏁 Завершена');



			if (canceledBookings.length === 0) {
				canceledList.innerHTML = '<p>Записей нет.</p>';
			} else {
				canceledBookings.forEach(booking => {
					const bookingCard = document.createElement('div');
					bookingCard.className = 'booking-card';
					bookingCard.innerHTML = `
			<div class="booking-info">
				<p><strong>ID записи:</strong> ${booking.id}</p>
				<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
				<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
				<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
				<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
				<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
				<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
				<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
				<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
				<p><strong>Статус:</strong> ${booking.status}</p>
			</div>
			`;
					canceledList.appendChild(bookingCard);
				});
			}
			if (workingBookings.length === 0) {
				workingList.innerHTML = '<p>Записей нет.</p>';
			} else {
				workingBookings.forEach(booking => {
					const bookingCard = document.createElement('div');
					bookingCard.className = 'booking-card';
					bookingCard.innerHTML = `
			<div class="booking-info">
				<p><strong>ID записи:</strong> ${booking.id}</p>
				<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
				<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
				<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
				<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
				<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
				<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
				<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
				<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
				<p><strong>Статус:</strong> ${booking.status}</p>
			</div>
	`;
					workingList.appendChild(bookingCard);
				});
			}
			if (endBookings.length === 0) {
				endList.innerHTML = '<p>Записей нет.</p>';
			} else {
				endBookings.forEach(booking => {
					const bookingCard = document.createElement('div');
					bookingCard.className = 'booking-card';
					bookingCard.innerHTML = `
			<div class="booking-info">
				<p><strong>ID записи:</strong> ${booking.id}</p>
				<p><strong>Название партнера:</strong> ${booking.partner_name}(${booking.partner_id})</p>
				<p><strong>Телефон партнера:</strong> ${booking.partner_phone}</p>
				<p><strong>Адрес партнера:</strong> ${booking.partner_address}</p>
				<p><strong>Имя пользователя:</strong> ${booking.user_name}(${booking.user_id})</p>
				<p><strong>Телефон пользователя:</strong> ${booking.user_phone}</p>
				<p><strong>Email пользователя:</strong> ${booking.user_email}</p>
				<p><strong>Дата бронирования:</strong> ${booking.booking_date}</p>
				<p><strong>Время бронирования:</strong> ${booking.booking_time}</p>
				<p><strong>Статус:</strong> ${booking.status}</p>
			</div>
	`;
					endList.appendChild(bookingCard);
				});
			}
			// Дополнительная проверка видимости после рендеринга
			const activeTab = document.querySelector('.order-tab-content.active');
			if (activeTab) {
				activeTab.style.display = 'block';
			}
		})
		.catch(error => {
			console.error('Ошибка загрузки записей:', error);
			const pendingList = document.getElementById('end-list');
			const confirmedList = document.getElementById('working-list');
			const canceledList = document.getElementById('canceled-list');
			if (pendingList) endList.innerHTML = '<p>Не удалось загрузить записи.</p>';
			if (confirmedList) workingListList.innerHTML = '<p>Не удалось загрузить записи.</p>';
			if (canceledList) canceledList.innerHTML = '<p>Не удалось загрузить записи.</p>';
		});
}