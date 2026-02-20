import { Navigate, Route, Routes } from "react-router-dom";
import { AppStateProvider, useAppState } from "./state/AppState";
import { AppLayout } from "./pages/AppLayout";
import { BookingPage } from "./pages/BookingPage";
import { DashboardPage } from "./pages/DashboardPage";
import { InventoryPage } from "./pages/InventoryPage";
import { LoginPage } from "./pages/LoginPage";
import { OrdersPage } from "./pages/OrdersPage";

function ProtectedRoutes() {
  const { session } = useAppState();
  if (!session) return <Navigate to="/login" replace />;
  return <AppLayout />;
}

export function App() {
  return (
    <AppStateProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/app" element={<ProtectedRoutes />}>
          <Route path="dashboard" element={<DashboardPage />} />
          <Route path="booking" element={<BookingPage />} />
          <Route path="inventory" element={<InventoryPage />} />
          <Route path="orders" element={<OrdersPage />} />
          <Route index element={<Navigate to="/app/dashboard" replace />} />
        </Route>
        <Route path="*" element={<Navigate to="/app/dashboard" replace />} />
      </Routes>
    </AppStateProvider>
  );
}


