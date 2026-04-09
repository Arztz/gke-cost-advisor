# =============================================================================
# Builder Stage - Build the Next.js application
# =============================================================================
FROM node:20-alpine AS builder

# Install pnpm for faster package management
RUN corepack enable && corepack prepare pnpm@latest --activate

# Set working directory
WORKDIR /app

# Copy package files first for dependency caching
COPY package.json pnpm-lock.yaml* ./

# Install dependencies
RUN if [ -f pnpm-lock.yaml ]; then \
      pnpm install --frozen-lockfile; \
    else \
      pnpm install; \
    fi

# Copy the rest of the application source code
COPY . .

# Build the Next.js application
RUN pnpm build

# =============================================================================
# Production Stage - Minimal production image
# =============================================================================
FROM node:20-alpine AS production

# Create non-root user and group
RUN addgroup -g 1000 -S nodejs && \
    adduser -u 1000 -S nodejs -G nodejs

# Set working directory
WORKDIR /app

# Copy the built application from builder stage
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

# Copy package.json for dependencies
COPY --from=builder /app/package.json ./

# Install only production dependencies
# The standalone output includes the node_modules needed to run
# But we need to ensure production deps are available
RUN corepack enable && corepack prepare pnpm@latest --activate && \
    pnpm install --prod --frozen-lockfile || true

# Change ownership to non-root user
RUN chown -R nodejs:nodejs /app

# Switch to non-root user
USER nodejs

# Expose the Next.js port
EXPOSE 3000

# Set environment variables
ENV NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

# Start the Next.js server
CMD ["node", "server.js"]