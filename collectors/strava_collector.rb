class StravaCollector
  def self.collect(token)
    client = Strava::Api::V3::Client.new(access_token: token)

    keys = %w(name distance moving_time location_city start_date)
    activity = client.list_athlete_activities.first.select do |key,_|
      keys.include? key
    end
    activity['created_at'] = Time.parse(activity['start_date'])
    activity.delete('start_date')
    activity['location'] = activity['location_city']
    activity.delete('location_city')
    return activity
  end
end
