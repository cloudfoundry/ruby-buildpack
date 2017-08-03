class TasksController < ApplicationController
  def index
    @tasks = Task.all

    respond_to do |format|
      format.html # index.html.erb
      format.json { render json: @tasks }

      # Example: Basic Usage
      format.pdf { render_task_list(@tasks) }
    end
  end

  def show
    @task = Task.find(params[:id])

    respond_to do |format|
      format.html # show.html.erb
      format.json { render json: @task }

      # Example: Using thinreports-rails gem
      # see https://github.com/takeshinoda/thinreports-rails
      format.pdf {
        send_data render_to_string, filename: "task#{@task.id}.pdf", 
                                    type: 'application/pdf', 
                                    disposition: 'inline'
      }
    end
  end

  def new
    @task = Task.new

    respond_to do |format|
      format.html # new.html.erb
      format.json { render json: @task }
    end
  end

  def edit
    @task = Task.find(params[:id])
  end

  def create
    @task = Task.new(params[:task])

    respond_to do |format|
      if @task.save
        format.html { redirect_to @task, notice: 'Task was successfully created.' }
        format.json { render json: @task, status: :created, location: @task }
      else
        format.html { render action: "new" }
        format.json { render json: @task.errors, status: :unprocessable_entity }
      end
    end
  end

  def update
    @task = Task.find(params[:id])

    respond_to do |format|
      if @task.update_attributes(params[:task])
        format.html { redirect_to @task, notice: 'Task was successfully updated.' }
        format.json { head :no_content }
      else
        format.html { render action: "edit" }
        format.json { render json: @task.errors, status: :unprocessable_entity }
      end
    end
  end

  def destroy
    @task = Task.find(params[:id])
    @task.destroy

    respond_to do |format|
      format.html { redirect_to tasks_url }
      format.json { head :no_content }
    end
  end

  private

  def render_task_list(tasks)
    report = ThinReports::Report.new layout: File.join(Rails.root, 'app', 'reports', 'tasks.tlf')

    tasks.each do |task|
      report.list.add_row do |row|
        row.values no: task.id, 
                   name: task.name, 
                   due_date: task.due_date, 
                   state: task.state
        row.item(:name).style(:color, 'red') unless task.done?
      end
    end
    
    send_data report.generate, filename: 'tasks.pdf', 
                               type: 'application/pdf', 
                               disposition: 'attachment'
  end
end
