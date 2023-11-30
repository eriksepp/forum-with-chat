// Combined throttle and debounce (trailing edge) function
export const throttleAndDebounce = (func, delay) => {
    let lastCall = 0;
    let debounceTimer = null;

    return function() {
        const now = new Date().getTime();
        const context = this;
        const args = arguments;

        clearTimeout(debounceTimer);

        // Debounce with trailing edge
        debounceTimer = setTimeout(() => {
            func.apply(context, args);
        }, delay);

        // Throttle
        if (now - lastCall >= delay) {
            lastCall = now;
            return func.apply(context, args);
        }
    };
};

// Func to get current date as ISO to sent to server
export const getCurrentISODate = () => {
    return new Date().toISOString();
}
