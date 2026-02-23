# frozen_string_literal: true
module TestRailsApp
  class Application < Rails::Application
    config.api_only = true
  end
end
