import { useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { addNotification, fetchUnreadCount } from '../redux/slices/notificationSlice';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

export function useNotificationStream() {
  const dispatch = useDispatch();
  const eventSourceRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const isConnectingRef = useRef(false);
  const { isAuthenticated } = useSelector((state) => state.auth);

  useEffect(() => {
    if (!isAuthenticated) {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      isConnectingRef.current = false;
      return;
    }

    const token = localStorage.getItem('accessToken');
    if (!token || isConnectingRef.current || eventSourceRef.current?.readyState === EventSource.OPEN) {
      return;
    }

    const connect = () => {
      if (isConnectingRef.current) return;
      isConnectingRef.current = true;

      const currentToken = localStorage.getItem('accessToken');
      if (!currentToken) {
        isConnectingRef.current = false;
        return;
      }

      const url = `${API_URL}/notifications/stream?token=${encodeURIComponent(currentToken)}`;
      const es = new EventSource(url);

      es.addEventListener('connected', (event) => {
        console.log('SSE connected:', JSON.parse(event.data));
      });

      es.addEventListener('notification', (event) => {
        try {
          const notification = JSON.parse(event.data);
          dispatch(addNotification(notification));

          if (Notification.permission === 'granted') {
            new Notification(notification.title, {
              body: notification.message,
              icon: '/favicon.svg',
            });
          }
        } catch (error) {
          console.error('Failed to parse notification:', error);
        }
      });

      es.addEventListener('ping', () => {});

      es.onopen = () => {
        isConnectingRef.current = false;
      };

      es.onerror = () => {
        es.close();
        eventSourceRef.current = null;
        isConnectingRef.current = false;
        reconnectTimeoutRef.current = setTimeout(connect, 5000);
      };

      eventSourceRef.current = es;
    };

    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      isConnectingRef.current = false;
    };
  }, [isAuthenticated, dispatch]);

  useEffect(() => {
    if (isAuthenticated) {
      dispatch(fetchUnreadCount());
      if (Notification.permission === 'default') {
        Notification.requestPermission();
      }
    }
  }, [isAuthenticated, dispatch]);

  return {};
}
