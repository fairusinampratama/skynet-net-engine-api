import { useState } from "react";
import useSWR, { mutate } from "swr"; // Added mutate
import { api, isolateUser } from "../../api/client"; // Added isolateUser
import { Wifi, ChevronLeft, ChevronRight, ArrowUpDown, ArrowUp, ArrowDown, Users, Lock, Shield, ShieldAlert, Ban, UserPlus } from "lucide-react"; // Added icons
import { Skeleton } from "../../components/ui/Skeleton";
import { AddUserModal } from "./AddUserModal";

const fetcher = (url) => api.get(url).then((res) => res.data);

export const CustomerTable = ({ routerId = 1, onSelectUser, selectedUser }) => {
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [sortColumn, setSortColumn] = useState('username');
    const [sortDirection, setSortDirection] = useState('asc');
    const [processingUser, setProcessingUser] = useState(null); // Track which user is being toggled
    const [isAddUserOpen, setIsAddUserOpen] = useState(false);

    const { data, error, isLoading, mutate } = useSWR(`/router/${routerId}/users`, fetcher, {
        refreshInterval: 10000,
        revalidateOnFocus: false,
    });

    const handleIsolate = async (e, user) => {
        e.stopPropagation();
        if (!user.ip) {
            alert("User has no IP, cannot isolate yet.");
            return;
        }

        const isIsolated = user.status === 'isolated';
        const action = isIsolated ? 'remove' : 'add';
        const confirmMsg = isIsolated
            ? `Restore access for ${user.username}?`
            : `ISOLATE ${user.username}? They will lose internet access.`;

        if (!window.confirm(confirmMsg)) return;

        setProcessingUser(user.username);
        try {
            // 1. Send Async Request (Returns immediately)
            await isolateUser({
                ip: user.ip,
                action: action,
                list: 'ISOLATED',
                router_id: routerId,
                comment: `Manual ${action} via Dashboard`
            }, false); // sync = false for robust async mode

            // 2. Poll for status change (Smart Polling)
            // We'll check every 2 seconds for up to 20 seconds
            let attempts = 0;
            const targetStatus = action === 'add' ? 'isolated' : 'connected'; // 'connected' or 'offline' really

            const pollInterval = setInterval(async () => {
                attempts++;
                const currentData = await mutate(); // Refresh SWR data
                const updatedUser = currentData.find(u => u.username === user.username);

                // Check if status matches expectation (or just changed from previous)
                // Simplify: just stop if status is what we want, or if we hit max attempts
                const isNowIsolated = updatedUser?.status === 'isolated';
                const success = (action === 'add' && isNowIsolated) || (action === 'remove' && !isNowIsolated);

                if (success || attempts >= 10) {
                    clearInterval(pollInterval);
                    setProcessingUser(null);
                    if (success) {
                        // Optional: Toast notification
                        console.log("Status updated successfully");
                    } else {
                        alert("Command sent, but status update is delayed. Check back shortly.");
                    }
                }
            }, 2000);

        } catch (err) {
            alert("Action failed: " + (err.response?.data?.error || err.message));
            setProcessingUser(null);
        }
    };



    // Sorting Logic
    const safeData = data || [];
    const sortedData = [...safeData].sort((a, b) => {
        let aVal = a[sortColumn] || '';
        let bVal = b[sortColumn] || '';

        // Handle status sorting (connected > isolated > offline)
        if (sortColumn === 'status') {
            const statusWeight = { connected: 0, isolated: 1, offline: 2 };
            aVal = statusWeight[a.status] ?? 3;
            bVal = statusWeight[b.status] ?? 3;
        }

        if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
        if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
        return 0;
    });

    // Pagination Logic
    const totalPages = Math.ceil(sortedData.length / pageSize);
    const startIdx = (page - 1) * pageSize;
    const endIdx = Math.min(startIdx + pageSize, sortedData.length);
    const paginatedData = sortedData.slice(startIdx, endIdx);

    // Sort handler
    const handleSort = (column) => {
        if (sortColumn === column) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortColumn(column);
            setSortDirection('asc');
        }
    };

    // Page size handler
    const handlePageSizeChange = (newSize) => {
        setPageSize(newSize);
        setPage(1); // Reset to first page
    };

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
            <div className="p-6 border-b border-slate-100 flex items-center justify-between">
                <h3 className="text-lg font-semibold text-slate-800 flex items-center gap-2">
                    <Users className="w-5 h-5 text-slate-500" />
                    Active Sessions
                </h3>
                <span className="text-sm font-medium text-slate-500">
                    Total: {sortedData.length}
                </span>
            </div>

            {/* Toolbar */}
            <div className="px-6 py-4 bg-slate-50 border-b border-slate-100 flex justify-end">
                <button
                    onClick={() => setIsAddUserOpen(true)}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm shadow-blue-600/20"
                >
                    <UserPlus className="w-4 h-4" />
                    Add Customer
                </button>
            </div>

            {isAddUserOpen && (
                <AddUserModal
                    routerId={routerId}
                    onClose={() => setIsAddUserOpen(false)}
                />
            )}

            <div className="overflow-x-auto">
                <table className="w-full text-left">
                    <thead className="bg-slate-50 border-b border-slate-200">
                        <tr>
                            <th
                                onClick={() => handleSort('username')}
                                className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider cursor-pointer hover:bg-slate-100 transition-colors"
                            >
                                <div className="flex items-center gap-1">
                                    User
                                    {sortColumn === 'username' ? (
                                        sortDirection === 'asc' ? <ArrowUp className="w-3 h-3" /> : <ArrowDown className="w-3 h-3" />
                                    ) : (
                                        <ArrowUpDown className="w-3 h-3 opacity-30" />
                                    )}
                                </div>
                            </th>
                            <th
                                onClick={() => handleSort('ip')}
                                className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider cursor-pointer hover:bg-slate-100 transition-colors"
                            >
                                <div className="flex items-center gap-1">
                                    IP Address
                                    {sortColumn === 'ip' ? (
                                        sortDirection === 'asc' ? <ArrowUp className="w-3 h-3" /> : <ArrowDown className="w-3 h-3" />
                                    ) : (
                                        <ArrowUpDown className="w-3 h-3 opacity-30" />
                                    )}
                                </div>
                            </th>
                            <th
                                onClick={() => handleSort('status')}
                                className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider cursor-pointer hover:bg-slate-100 transition-colors"
                            >
                                <div className="flex items-center gap-1">
                                    Status
                                    {sortColumn === 'status' ? (
                                        sortDirection === 'asc' ? <ArrowUp className="w-3 h-3" /> : <ArrowDown className="w-3 h-3" />
                                    ) : (
                                        <ArrowUpDown className="w-3 h-3 opacity-30" />
                                    )}
                                </div>
                            </th>
                            <th
                                onClick={() => handleSort('profile')}
                                className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider cursor-pointer hover:bg-slate-100 transition-colors"
                            >
                                <div className="flex items-center gap-1">
                                    Profile
                                    {sortColumn === 'profile' ? (
                                        sortDirection === 'asc' ? <ArrowUp className="w-3 h-3" /> : <ArrowDown className="w-3 h-3" />
                                    ) : (
                                        <ArrowUpDown className="w-3 h-3 opacity-30" />
                                    )}
                                </div>
                            </th>
                            <th className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                                Actions
                            </th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                        {isLoading && Array.from({ length: 5 }).map((_, i) => (
                            <tr key={i}>
                                <td className="px-6 py-4"><Skeleton className="h-4 w-32" /></td>
                                <td className="px-6 py-4"><Skeleton className="h-4 w-24" /></td>
                                <td className="px-6 py-4"><Skeleton className="h-6 w-20 rounded-full" /></td>
                                <td className="px-6 py-4"><Skeleton className="h-6 w-16 rounded-full" /></td>
                                <td className="px-6 py-4"><Skeleton className="h-8 w-8 rounded-lg" /></td>
                            </tr>
                        ))}

                        {paginatedData.map((user, i) => (
                            <tr
                                key={i}
                                onClick={() => onSelectUser && onSelectUser(user.username)}
                                className={`cursor-pointer transition-colors ${selectedUser === user.username ? 'bg-blue-50' : 'hover:bg-slate-50'}`}
                            >
                                <td className="px-6 py-4 font-medium text-slate-900 flex items-center gap-2">
                                    {selectedUser === user.username && <div className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />}
                                    {user.username}
                                </td>
                                <td className="px-6 py-4 text-slate-600 font-mono text-sm">{user.ip || '-'}</td>
                                <td className="px-6 py-4">
                                    {user.status === 'online' ? (
                                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 border border-green-200">
                                            <Wifi className="w-3 h-3" />
                                            Online
                                        </span>
                                    ) : (
                                        <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600 border border-slate-200">
                                            <div className="w-2 h-2 rounded-full bg-slate-400" />
                                            Offline
                                        </span>
                                    )}
                                </td>
                                {/* Profile Badge Column */}
                                <td className="px-6 py-4">
                                    {user.profile ? (
                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${user.profile === 'isolirebilling'
                                            ? 'bg-red-50 text-red-700 border border-red-200'
                                            : user.profile.includes('100')
                                                ? 'bg-purple-50 text-purple-700 border border-purple-200'
                                                : user.profile.includes('50')
                                                    ? 'bg-blue-50 text-blue-700 border border-blue-200'
                                                    : user.profile.includes('10')
                                                        ? 'bg-cyan-50 text-cyan-700 border border-cyan-200'
                                                        : 'bg-slate-50 text-slate-600 border border-slate-200'
                                            }`}>
                                            {user.profile}
                                        </span>
                                    ) : (
                                        <span className="text-slate-400 text-xs">-</span>
                                    )}
                                </td>
                                <td className="px-6 py-4">
                                    <button
                                        onClick={(e) => handleIsolate(e, user)}
                                        disabled={processingUser === user.username}
                                        className={`p-2 rounded-lg border transition-all ${user.status === 'isolated'
                                            ? 'bg-green-50 text-green-600 border-green-200 hover:bg-green-100'
                                            : 'bg-red-50 text-red-600 border-red-200 hover:bg-red-100'
                                            }`}
                                        title={user.status === 'isolated' ? "Restore Connection" : "Isolate User"}
                                    >
                                        {processingUser === user.username ? (
                                            <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
                                        ) : user.status === 'isolated' ? (
                                            <Shield className="w-4 h-4" />
                                        ) : (
                                            <Ban className="w-4 h-4" />
                                        )}
                                    </button>
                                </td>
                            </tr>
                        ))}

                        {!isLoading && sortedData.length === 0 && (
                            <tr><td colSpan="5" className="px-6 py-8 text-center text-slate-400">No active users found.</td></tr>
                        )}
                    </tbody>
                </table>
            </div>

            {/* Advanced Pagination Controls */}
            <div className="flex items-center justify-between px-6 py-4 border-t border-slate-200 bg-slate-50">
                <div className="flex items-center gap-3">
                    <span className="text-sm text-slate-600 font-medium">
                        Showing <span className="text-slate-900">{startIdx + 1}-{endIdx}</span> of <span className="text-slate-900">{sortedData.length}</span>
                    </span>
                    <div className="flex items-center gap-2">
                        <span className="text-sm text-slate-500">Show:</span>
                        <select
                            value={pageSize}
                            onChange={(e) => handlePageSizeChange(Number(e.target.value))}
                            className="px-3 py-1.5 text-sm font-medium border border-slate-300 rounded-lg bg-white hover:border-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent cursor-pointer transition-colors"
                        >
                            <option value={10}>10</option>
                            <option value={25}>25</option>
                            <option value={50}>50</option>
                            <option value={100}>100</option>
                        </select>
                        <span className="text-sm text-slate-500">entries</span>
                    </div>
                </div>

                <div className="flex items-center gap-2">
                    <button
                        onClick={() => setPage(1)}
                        disabled={page === 1}
                        className="px-2 py-1 text-sm font-medium text-slate-600 hover:text-slate-900 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        ««
                    </button>
                    <button
                        onClick={() => setPage(p => Math.max(1, p - 1))}
                        disabled={page === 1}
                        className="px-3 py-1 text-sm font-medium text-slate-600 hover:text-slate-900 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                    >
                        <ChevronLeft className="w-4 h-4" />
                        Previous
                    </button>

                    <span className="text-sm text-slate-600 px-2">
                        Page {page} of {totalPages}
                    </span>

                    <button
                        onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                        disabled={page === totalPages}
                        className="px-3 py-1 text-sm font-medium text-slate-600 hover:text-slate-900 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                    >
                        Next
                        <ChevronRight className="w-4 h-4" />
                    </button>
                    <button
                        onClick={() => setPage(totalPages)}
                        disabled={page === totalPages}
                        className="px-2 py-1 text-sm font-medium text-slate-600 hover:text-slate-900 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        »»
                    </button>
                </div>
            </div>
        </div>
    );
};
