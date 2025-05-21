document.addEventListener('DOMContentLoaded', () => {
	loadBookings()
	// Загрузка записей для дашборда
});

// Обновление статуса записи
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
				alert(`Запись успешно ${status === 'confirmed' ? 'подтверждена' : 'отменена'}!`);
				loadBookings();
			}
		})
		.catch(error => {
			console.error('Ошибка обновления статуса:', error);
			alert('Не удалось обновить статус записи.');
		});
}

function loadBookings() {
	fetch('/api/bookings')
		.then(response => response.json())
		.then(bookings => {
			const bookingsList = document.getElementById('bookings-list');
			bookingsList.innerHTML = ' ';

			if (bookings.length === 0) {
				bookingsList.innerHTML = '<p>Записей нет.</p>';
				return;
			}

			bookings.forEach(booking => {
				const bookingCard = document.createElement('div');
				bookingCard.className = 'booking-card';
				bookingCard.innerHTML = `
				<div class="booking-info">
					<p><strong>ID пользователя:</strong> ${booking.user_id}</p>
					<p><strong>Дата:</strong> ${booking.booking_date}</p>
					<p><strong>Время:</strong> ${booking.booking_time}</p>
					<p><strong>Статус:</strong> ${booking.status}</p>
				</div>
				<div class="booking-actions">
					<button onclick="updateBookingStatus(${booking.id}, 'confirmed')">Подтвердить</button>
					<button onclick="updateBookingStatus(${booking.id}, 'canceled')">Отменить</button>
				</div>
			`;
				bookingsList.appendChild(bookingCard);
			});
		})
		.catch(error => {
			console.error('Ошибка загрузки записей:', error);
			document.getElementById('bookings-list').innerHTML = '<p>Не удалось загрузить записи.</p>';
		});
}